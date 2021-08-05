// RAINBOND, Application Management Platform
// Copyright (C) 2020-2021 Goodrain Co., Ltd.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. For any non-GPL usage of Rainbond,
// one or multiple Commercial Licenses authorized by Goodrain Co., Ltd.
// must be obtained first.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package operator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	rainbondv1alpha1 "github.com/goodrain/rainbond-operator/api/v1alpha1"
	"github.com/goodrain/rainbond-operator/util/commonutil"
	"github.com/goodrain/rainbond-operator/util/constants"
	"github.com/goodrain/rainbond-operator/util/rbdutil"
	"github.com/goodrain/rainbond-operator/util/retryutil"
	"github.com/goodrain/rainbond-operator/util/suffixdomain"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"goodrain.com/cloud-adaptor/internal/adaptor/v1alpha1"
	"goodrain.com/cloud-adaptor/internal/repo"
	"goodrain.com/cloud-adaptor/version"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var helmPath = "/Users/barnett/bin/helm"
var chartPath = "/Users/barnett/coding/gopath/src/goodrain.com/cloud-adaptor/chart"

func init() {
	if os.Getenv("HELM_PATH") != "" {
		helmPath = os.Getenv("HELM_PATH")
	}
	if os.Getenv("CHART_PATH") != "" {
		chartPath = os.Getenv("CHART_PATH")
	}
}

//RainbondRegionInit rainbond region init by operator
type RainbondRegionInit struct {
	kubeconfig                v1alpha1.KubeConfig
	namespace                 string
	rainbondClusterConfigRepo repo.RainbondClusterConfigRepository
}

//NewRainbondRegionInit new
func NewRainbondRegionInit(kubeconfig v1alpha1.KubeConfig, rainbondClusterConfigRepo repo.RainbondClusterConfigRepository) *RainbondRegionInit {
	return &RainbondRegionInit{
		kubeconfig:                kubeconfig,
		namespace:                 constants.Namespace,
		rainbondClusterConfigRepo: rainbondClusterConfigRepo,
	}
}

//InitRainbondRegion init rainbond region
func (r *RainbondRegionInit) InitRainbondRegion(initConfig *v1alpha1.RainbondInitConfig) error {
	clusterID := initConfig.ClusterID
	kubeconfigFileName := "/tmp/" + clusterID + ".kubeconfig"
	if err := r.kubeconfig.Save(kubeconfigFileName); err != nil {
		return fmt.Errorf("warite kubeconfig file failure %s", err.Error())
	}
	defer func() {
		os.Remove(kubeconfigFileName)
	}()
	// create namespace
	client, runtimeClient, err := r.kubeconfig.GetKubeClient()
	if err != nil {
		return fmt.Errorf("create kube client failure %s", err.Error())
	}
	cn := &v1.Namespace{}
	cn.Name = r.namespace
	if err := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		_, err = client.CoreV1().Namespaces().Create(ctx, cn, metav1.CreateOptions{})
		if err != nil && !k8sErrors.IsAlreadyExists(err) {
			return fmt.Errorf("create namespace failure %s", err.Error())
		}
		return nil
	}(); err != nil {
		return err
	}

	// make sure ClusterRoleBinding rainbond-operator not exists.
	if err := r.ensureClusterRoleBinding(client); err != nil {
		return errors.WithMessage(err, "ensure clusterrolebinding rainbond-operator")
	}

	// helm create rainbond operator chart
	defaultArgs := []string{
		helmPath, "install", "rainbond-operator", chartPath, "-n", r.namespace,
		"--kubeconfig", kubeconfigFileName,
		"--set", "operator.image.name=" + fmt.Sprintf("%s/rainbond-operator", version.InstallImageRepo),
		"--set", "operator.image.tag=" + version.OperatorVersion}
	logrus.Infof(strings.Join(defaultArgs, " "))
	for {
		var stdout = bytes.NewBuffer(nil)
		cmd := &exec.Cmd{
			Path:   helmPath,
			Args:   defaultArgs,
			Stdout: stdout,
			Stdin:  os.Stdin,
			Stderr: stdout,
		}
		if err := cmd.Run(); err != nil {
			errout := stdout.String()
			if !strings.Contains(errout, "cannot re-use a name that is still in use") {
				if strings.Contains(errout, "\"rainbond-operator\" already exists") {
					func() {
						ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
						defer cancel()
						client.RbacV1().ClusterRoleBindings().Delete(ctx, "rainbond-operator", metav1.DeleteOptions{})
					}()
					continue
				}
				return fmt.Errorf("install chart failure %s, %s", err.Error(), errout)
			}
			logrus.Warning("rainbond operator chart release is exist")
		}
		break
	}
	// waiting operator is ready
	ticker := time.NewTicker(time.Second * 5)
	timer := time.NewTimer(time.Minute * 10)
	defer timer.Stop()
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-timer.C:
			return fmt.Errorf("waiting rainbond operator ready timeout")
		}
		var rb *rbacv1.ClusterRoleBinding
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			rb, err = client.RbacV1().ClusterRoleBindings().Get(ctx, "rainbond-operator", metav1.GetOptions{})
			if err != nil {
				logrus.Errorf("get role binding rainbond-operator status failure %s", err.Error())
			}
		}()
		if rb != nil && rb.Name == "rainbond-operator" {
			break
		}
	}
	// create custom resource
	if err := r.createRainbondCR(client, runtimeClient, initConfig); err != nil {
		return fmt.Errorf("create rainbond CR failure %s", err.Error())
	}
	return nil
}

func (r *RainbondRegionInit) createRainbondCR(kubeClient *kubernetes.Clientset, client client.Client, initConfig *v1alpha1.RainbondInitConfig) error {
	// create rainbond cluster resource
	//TODO: define etcd config by RainbondInitConfig
	rcc, err := r.rainbondClusterConfigRepo.Get(initConfig.ClusterID)
	if err != nil && err != gorm.ErrRecordNotFound {
		logrus.Errorf("get rainbond cluster config failure %s", err.Error())
	}
	cluster := &rainbondv1alpha1.RainbondCluster{
		Spec: rainbondv1alpha1.RainbondClusterSpec{
			InstallVersion:          initConfig.RainbondVersion,
			CIVersion:               initConfig.RainbondCIVersion,
			EnableHA:                initConfig.EnableHA,
			RainbondImageRepository: version.InstallImageRepo,
			SuffixHTTPHost:          initConfig.SuffixHTTPHost,
			NodesForChaos:           initConfig.ChaosNodes,
			NodesForGateway:         initConfig.GatewayNodes,
			GatewayIngressIPs:       initConfig.EIPs,
		},
	}
	if rcc != nil {
		logrus.Info("use custom rainbondcluster config")
		if err := yaml.Unmarshal([]byte(rcc.Config), cluster); err != nil {
			logrus.Errorf("Unmarshal rainbond config failure %s", err.Error())
		}
	}
	if len(cluster.Spec.GatewayIngressIPs) == 0 {
		return fmt.Errorf("can not select eip, please specify `gatewayIngressIPs` in the custom cluster init configuration")
	}
	if cluster.Spec.EtcdConfig != nil && len(cluster.Spec.EtcdConfig.Endpoints) == 0 {
		cluster.Spec.EtcdConfig = nil
	}
	cluster.Spec.InstallMode = "WithoutPackage"
	// default build cache mode set is `hostpath`
	if cluster.Spec.CacheMode == "" {
		cluster.Spec.CacheMode = "hostpath"
	}

	cluster.Spec.ConfigCompleted = true
	// image hub must be nil, where not define
	if cluster.Spec.ImageHub != nil && cluster.Spec.ImageHub.Domain == "" {
		cluster.Spec.ImageHub = nil
	}
	if cluster.Spec.InstallVersion == "" {
		cluster.Spec.InstallVersion = initConfig.RainbondVersion
	}
	if cluster.Spec.CIVersion == "" {
		cluster.Spec.CIVersion = initConfig.RainbondCIVersion
	}
	if cluster.Spec.RainbondImageRepository == "" {
		cluster.Spec.RainbondImageRepository = version.InstallImageRepo
	}
	if initConfig.ETCDConfig != nil && len(initConfig.ETCDConfig.Endpoints) > 0 {
		cluster.Spec.EtcdConfig = initConfig.ETCDConfig
	}
	if initConfig.RegionDatabase != nil && initConfig.RegionDatabase.Host != "" {
		cluster.Spec.RegionDatabase = &rainbondv1alpha1.Database{
			Host:     initConfig.RegionDatabase.Host,
			Port:     initConfig.RegionDatabase.Port,
			Username: initConfig.RegionDatabase.UserName,
			Password: initConfig.RegionDatabase.Password,
		}
	}
	if initConfig.NasServer != "" {
		cluster.Spec.RainbondVolumeSpecRWX = &rainbondv1alpha1.RainbondVolumeSpec{
			CSIPlugin: &rainbondv1alpha1.CSIPluginSource{
				AliyunNas: &rainbondv1alpha1.AliyunNasCSIPluginSource{
					AccessKeyID:     "",
					AccessKeySecret: "",
				},
			},
			StorageClassParameters: &rainbondv1alpha1.StorageClassParameters{
				Parameters: map[string]string{
					"volumeAs":        "subpath",
					"server":          initConfig.NasServer,
					"archiveOnDelete": "true",
				},
			},
		}
	}
	// handle volume spec
	if cluster.Spec.RainbondVolumeSpecRWX != nil {
		if cluster.Spec.RainbondVolumeSpecRWX.CSIPlugin != nil {
			if cluster.Spec.RainbondVolumeSpecRWX.CSIPlugin.AliyunCloudDisk == nil &&
				cluster.Spec.RainbondVolumeSpecRWX.CSIPlugin.AliyunNas == nil &&
				cluster.Spec.RainbondVolumeSpecRWX.CSIPlugin.NFS == nil {
				cluster.Spec.RainbondVolumeSpecRWX.CSIPlugin = nil
			}
		}
	}
	if cluster.Spec.RainbondVolumeSpecRWO != nil {
		if cluster.Spec.RainbondVolumeSpecRWO.CSIPlugin != nil {
			if cluster.Spec.RainbondVolumeSpecRWO.CSIPlugin.AliyunCloudDisk == nil &&
				cluster.Spec.RainbondVolumeSpecRWO.CSIPlugin.AliyunNas == nil &&
				cluster.Spec.RainbondVolumeSpecRWO.CSIPlugin.NFS == nil {
				cluster.Spec.RainbondVolumeSpecRWO.CSIPlugin = nil
			}
		}
		if cluster.Spec.RainbondVolumeSpecRWO.CSIPlugin == nil && cluster.Spec.RainbondVolumeSpecRWO.StorageClassName == "" {
			cluster.Spec.RainbondVolumeSpecRWO = nil
		}
	}
	if cluster.Spec.RainbondVolumeSpecRWX == nil ||
		(cluster.Spec.RainbondVolumeSpecRWX.CSIPlugin == nil &&
			cluster.Spec.RainbondVolumeSpecRWX.StorageClassName == "") {
		cluster.Spec.RainbondVolumeSpecRWX = &rainbondv1alpha1.RainbondVolumeSpec{
			CSIPlugin: &rainbondv1alpha1.CSIPluginSource{
				NFS: &rainbondv1alpha1.NFSCSIPluginSource{},
			},
		}
	}
	if cluster.Spec.SuffixHTTPHost == "" {
		var ip string
		if len(initConfig.GatewayNodes) > 0 {
			ip = initConfig.GatewayNodes[0].InternalIP
		}
		if len(initConfig.EIPs) > 0 && initConfig.EIPs[0] != "" {
			ip = initConfig.EIPs[0]
		}
		if ip != "" {
			err := retryutil.Retry(1*time.Second, 3, func() (bool, error) {
				domain, err := r.genSuffixHTTPHost(kubeClient, ip)
				if err != nil {
					return false, err
				}
				cluster.Spec.SuffixHTTPHost = domain
				return true, nil
			})
			if err != nil {
				logrus.Warningf("generate suffix http host: %v", err)
				cluster.Spec.SuffixHTTPHost = constants.DefHTTPDomainSuffix
			}
		} else {
			cluster.Spec.SuffixHTTPHost = constants.DefHTTPDomainSuffix
		}
	}
	cluster.Name = "rainbondcluster"
	cluster.Namespace = r.namespace
	operator, err := NewOperator(Config{
		RainbondVersion:         initConfig.RainbondVersion,
		Namespace:               r.namespace,
		ArchiveFilePath:         "/opt/rainbond/pkg/tgz/rainbond.tgz",
		RuntimeClient:           client,
		Rainbondpackage:         "rainbondpackage",
		RainbondImageRepository: version.InstallImageRepo,
		OnlyInstallRegion:       true,
	})
	if err != nil {
		return fmt.Errorf("create operator instance failure %s", err.Error())
	}
	return operator.Install(cluster)
}

func (r *RainbondRegionInit) genSuffixHTTPHost(kubeClient *kubernetes.Clientset, ip string) (domain string, err error) {
	id, auth, err := r.getOrCreateUUIDAndAuth(kubeClient)
	if err != nil {
		return "", err
	}
	domain, err = suffixdomain.GenerateDomain(ip, id, auth)
	if err != nil {
		return "", err
	}
	return domain, nil
}

func (r *RainbondRegionInit) getOrCreateUUIDAndAuth(kubeClient *kubernetes.Clientset) (id, auth string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	cm, err := kubeClient.CoreV1().ConfigMaps(r.namespace).Get(ctx, "rbd-suffix-host", metav1.GetOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return "", "", err
	}
	if k8sErrors.IsNotFound(err) {
		logrus.Info("not found configmap rbd-suffix-host, create it")
		cm = generateSuffixConfigMap("rbd-suffix-host", r.namespace)
		if _, err = kubeClient.CoreV1().ConfigMaps(r.namespace).Create(ctx, cm, metav1.CreateOptions{}); err != nil {
			return "", "", err
		}

	}
	return cm.Data["uuid"], cm.Data["auth"], nil
}

func generateSuffixConfigMap(name, namespace string) *v1.ConfigMap {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string]string{
			"uuid": string(uuid.NewUUID()),
			"auth": string(uuid.NewUUID()),
		},
	}
	return cm
}

//GetRainbondRegionStatus get rainbond region status
func (r *RainbondRegionInit) GetRainbondRegionStatus(clusterID string) (*v1alpha1.RainbondRegionStatus, error) {
	coreClient, rainbondClient, err := r.kubeconfig.GetKubeClient()
	if err != nil {
		return nil, err
	}
	status := &v1alpha1.RainbondRegionStatus{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	deployment, err := coreClient.AppsV1().Deployments("rbd-system").Get(ctx, "rainbond-operator", metav1.GetOptions{})
	if err != nil {
		logrus.Warningf("get operator failure %s", err.Error())
	}
	if deployment != nil && deployment.Status.ReadyReplicas >= 1 {
		status.OperatorReady = true
	}
	var cluster rainbondv1alpha1.RainbondCluster
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel2()
	err = rainbondClient.Get(ctx2, types.NamespacedName{Name: "rainbondcluster", Namespace: "rbd-system"}, &cluster)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, err
		}
		logrus.Warningf("get cluster failure %s", err.Error())
	}
	status.RainbondCluster = &cluster
	var pkgStatus rainbondv1alpha1.RainbondPackage
	ctx3, cancel3 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel3()
	err = rainbondClient.Get(ctx3, types.NamespacedName{Name: "rainbondpackage", Namespace: "rbd-system"}, &pkgStatus)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, err
		}
		logrus.Warningf("get pkgStatus failure %s", err.Error())
	}
	status.RainbondPackage = &pkgStatus
	var volume rainbondv1alpha1.RainbondVolume
	ctx4, cancel4 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel4()
	err = rainbondClient.Get(ctx4, types.NamespacedName{Name: "rainbondvolumerwx", Namespace: "rbd-system"}, &volume)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, err
		}
		logrus.Warningf("get rainbond volume failure %s", err.Error())
	}
	status.RainbondVolume = &volume
	ctx5, cancel5 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel5()
	config, err := coreClient.CoreV1().ConfigMaps("rbd-system").Get(ctx5, "region-config", metav1.GetOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		logrus.Warningf("get region config failure %s", err.Error())
	}
	status.RegionConfig = config
	return status, nil
}

//UninstallRegion uninstall
func (r *RainbondRegionInit) UninstallRegion(clusterID string) error {
	deleteOpts := metav1.DeleteOptions{
		GracePeriodSeconds: commonutil.Int64(0),
	}
	coreClient, runtimeClient, err := r.kubeconfig.GetKubeClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	// delte rainbond components
	if err := runtimeClient.DeleteAllOf(ctx, &rainbondv1alpha1.RbdComponent{}, client.InNamespace(r.namespace)); err != nil {
		return fmt.Errorf("delete component failure: %v", err)
	}
	// delete rainbond packages
	if err := runtimeClient.DeleteAllOf(ctx, &rainbondv1alpha1.RainbondPackage{}, client.InNamespace(r.namespace)); err != nil {
		return fmt.Errorf("delete rainbond package failure: %v", err)
	}
	// delete rainbondvolume
	if err := runtimeClient.DeleteAllOf(ctx, &rainbondv1alpha1.RainbondVolume{}, client.InNamespace(r.namespace)); err != nil {
		return fmt.Errorf("delete rainbond volume failure: %v", err)
	}

	// delete pv based on pvc
	claims, err := coreClient.CoreV1().PersistentVolumeClaims(r.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list pv: %v", err)
	}
	for _, claim := range claims.Items {
		if claim.Spec.VolumeName == "" {
			// unbound pvc
			continue
		}
		if err := coreClient.CoreV1().PersistentVolumes().Delete(ctx, claim.Spec.VolumeName, metav1.DeleteOptions{}); err != nil {
			if k8sErrors.IsNotFound(err) {
				continue
			}
			return fmt.Errorf("delete persistent volume: %v", err)
		}
	}
	// delete pvc
	if err := coreClient.CoreV1().PersistentVolumeClaims(r.namespace).DeleteCollection(ctx, deleteOpts, metav1.ListOptions{}); err != nil {
		return fmt.Errorf("delete persistent volume claims: %v", err)
	}

	// delete storage class and csidriver
	rainbondLabelSelector := fields.SelectorFromSet(rbdutil.LabelsForRainbond(nil)).String()
	if err := coreClient.StorageV1().StorageClasses().DeleteCollection(ctx, deleteOpts, metav1.ListOptions{LabelSelector: rainbondLabelSelector}); err != nil {
		return fmt.Errorf("delete storageclass: %v", err)
	}
	if err := coreClient.StorageV1().StorageClasses().Delete(ctx, "rainbondslsc", metav1.DeleteOptions{}); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return fmt.Errorf("delete storageclass rainbondslsc: %v", err)
		}
	}
	if err := coreClient.StorageV1().StorageClasses().Delete(ctx, "rainbondsssc", metav1.DeleteOptions{}); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return fmt.Errorf("delete storageclass rainbondsssc: %v", err)
		}
	}
	if err := coreClient.StorageV1beta1().CSIDrivers().DeleteCollection(ctx, deleteOpts, metav1.ListOptions{LabelSelector: rainbondLabelSelector}); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return fmt.Errorf("delete csidriver: %v", err)
		}
	}

	// delete rainbond-operator ClusterRoleBinding
	if err := coreClient.RbacV1().ClusterRoleBindings().Delete(ctx, "rainbond-operator", metav1.DeleteOptions{}); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return fmt.Errorf("delete cluster role bindings: %v", err)
		}
	}

	// delete rainbond cluster
	var rbdcluster rainbondv1alpha1.RainbondCluster
	if err := runtimeClient.DeleteAllOf(ctx, &rbdcluster, client.InNamespace(r.namespace)); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return fmt.Errorf("delete rainbond volume failure: %v", err)
		}
	}

	if err := coreClient.CoreV1().Namespaces().Delete(ctx, r.namespace, metav1.DeleteOptions{}); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return fmt.Errorf("delete namespace %s failure: %v", r.namespace, err)
		}
	}
	ticker := time.NewTicker(time.Second * 5)
	timer := time.NewTimer(time.Minute * 10)
	defer timer.Stop()
	defer ticker.Stop()
	for {
		if _, err := coreClient.CoreV1().Namespaces().Get(ctx, r.namespace, metav1.GetOptions{}); err != nil {
			if k8sErrors.IsNotFound(err) {
				return nil
			}
		}
		select {
		case <-timer.C:
			return fmt.Errorf("waiting namespace deleted timeout")
		case <-ticker.C:
			logrus.Debugf("waiting namespace rbd-system deleted")
		}
	}
}

func (r *RainbondRegionInit) ensureClusterRoleBinding(kubeClient kubernetes.Interface) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	crb, err := kubeClient.RbacV1().ClusterRoleBindings().Get(ctx, "rainbond-operator", metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if err := kubeClient.RbacV1().ClusterRoleBindings().Delete(ctx, crb.Name, metav1.DeleteOptions{}); err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}
