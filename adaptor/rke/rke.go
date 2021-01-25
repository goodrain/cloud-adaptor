package rke

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	rainbondv1alpha1 "github.com/goodrain/rainbond-operator/api/v1alpha1"
	"github.com/jinzhu/gorm"
	"github.com/rancher/rke/cluster"
	"github.com/rancher/rke/cmd"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/log"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/pki/cert"
	v3 "github.com/rancher/rke/types"
	"github.com/sirupsen/logrus"
	"goodrain.com/cloud-adaptor/adaptor"
	"goodrain.com/cloud-adaptor/adaptor/v1alpha1"
	"goodrain.com/cloud-adaptor/api/infrastructure/datastore"
	"goodrain.com/cloud-adaptor/api/models"
	"goodrain.com/cloud-adaptor/library/bcode"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
)

type rkeAdaptor struct {
	Repo *ClusterRepo
}

//Create create ack adaptor
func Create() (adaptor.RainbondClusterAdaptor, error) {
	return &rkeAdaptor{
		Repo: NewRKEClusterRepo(datastore.GetGDB()),
	}, nil
}

func toString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (r *rkeAdaptor) ClusterList() ([]*v1alpha1.Cluster, error) {
	rkeclusters, err := r.Repo.ListCluster()
	if err != nil {
		return nil, fmt.Errorf("get cluster meta info failure %s", err.Error())
	}
	var re []*v1alpha1.Cluster
	var wait sync.WaitGroup
	for _, rc := range rkeclusters {
		wait.Add(1)
		go func(rc *models.RKECluster) {
			defer wait.Done()
			re = append(re, converClusterMeta(rc))
		}(rc)
	}
	wait.Wait()
	return re, nil
}

func (r *rkeAdaptor) DescribeCluster(clusterID string) (*v1alpha1.Cluster, error) {
	rkecluster, err := r.Repo.GetCluster(clusterID)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s meta info failure %s", clusterID, err.Error())
	}
	return converClusterMeta(rkecluster), nil
}

func (r *rkeAdaptor) DeleteCluster(clusterID string) error {
	cluster, _ := r.DescribeCluster(clusterID)
	if cluster != nil && cluster.RainbondInit {
		return bcode.ErrClusterNotAllowDelete
	}
	return r.Repo.DeleteCluster(clusterID)
}

func (r *rkeAdaptor) GetRainbondInitConfig(cluster *v1alpha1.Cluster, gateway, chaos []*rainbondv1alpha1.K8sNode, rollback func(step, message, status string)) *v1alpha1.RainbondInitConfig {
	return &v1alpha1.RainbondInitConfig{
		EnableHA: func() bool {
			if cluster.Size > 3 {
				return true
			}
			return false
		}(),
		ClusterID:    cluster.ClusterID,
		GatewayNodes: gateway,
		ChaosNodes:   chaos,
		EIPs: func() (re []string) {
			if len(cluster.EIP) > 0 {
				return cluster.EIP
			}
			for _, n := range gateway {
				if n.ExternalIP != "" {
					re = append(re, n.ExternalIP)
				}
			}
			return
		}(),
	}
}

func (r *rkeAdaptor) CreateRainbondKubernetes(ctx context.Context, config *v1alpha1.KubernetesClusterConfig, rollback func(step, message, status string)) *v1alpha1.Cluster {
	rollback("InitClusterConfig", "", "start")
	rkecluster, err := r.Repo.GetCluster(config.ClusterName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			rkecluster = &models.RKECluster{
				Name:  config.ClusterName,
				Stats: v1alpha1.InitState,
			}
			r.Repo.Create(rkecluster)
		} else {
			logrus.Errorf("get cluster meta info failure %s", err.Error())
			rollback("InitClusterConfig", "Get cluster meta info failure", "failure")
			return nil
		}
	}

	if len(config.Nodes) < 0 {
		rollback("InitClusterConfig", "Provide at least one node", "failure")
		return nil
	}
	var masterNode, etcdNode, workerNode int
	for _, node := range config.Nodes {
		if strings.Contains(strings.Join(node.Roles, ","), "controlplane") {
			masterNode++
		}
		if strings.Contains(strings.Join(node.Roles, ","), "worker") {
			workerNode++
		}
		if strings.Contains(strings.Join(node.Roles, ","), "etcd") {
			etcdNode++
		}
	}
	config.WorkerNodeNum = workerNode
	config.MasterNodeNum = masterNode
	config.ETCDNodeNum = etcdNode
	if config.WorkerNodeNum == 0 {
		rollback("InitClusterConfig", "Provide at least one compute node", "failure")
		return nil
	}
	if config.MasterNodeNum == 0 {
		rollback("InitClusterConfig", "Provide at least one master node", "failure")
		return nil
	}
	if config.ETCDNodeNum == 0 {
		rollback("InitClusterConfig", "Provide at least one etcd node", "failure")
		return nil
	}
	clusterConfig := v1alpha1.GetDefaultRKECreateClusterConfig(*config)
	rkeConfig, _ := clusterConfig.(*v3.RancherKubernetesEngineConfig)

	// create rke cluster config
	configDir := "/tmp"
	if os.Getenv("CONFIG_DIR") != "" {
		configDir = os.Getenv("CONFIG_DIR")
	}
	clusterStatPath := fmt.Sprintf("%s/rke/%s", configDir, config.ClusterName)

	if rkecluster.Stats == v1alpha1.InstallFailed {
		// clear local state data
		os.RemoveAll(clusterStatPath)
		rkecluster.Stats = v1alpha1.InitState
		if err := r.Repo.Update(rkecluster); err != nil {
			logrus.Errorf("update rke cluster %s state failure %s", rkecluster.Name, err.Error())
		}
	}
	os.MkdirAll(clusterStatPath, 0755)

	filePath := fmt.Sprintf("%s/cluster.yml", clusterStatPath)
	out, _ := yaml.Marshal(rkeConfig)
	if err := ioutil.WriteFile(filePath, out, 0755); err != nil {
		rollback("InitClusterConfig", err.Error(), "failure")
		logrus.Errorf("write rke cluster config file failure %s", err.Error())
		return nil
	}

	// setting up the flags
	flags := cluster.GetExternalFlags(false, false, false, false, "", filePath)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// set install log out
	logPath := fmt.Sprintf("%s/create.log", clusterStatPath)
	writer, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		logrus.Errorf("open create cluster log file %s failure %s", logPath, err.Error())
	}
	logger := logrus.New()
	if writer != nil {
		defer writer.Close()
		logger.Out = writer
	}
	ctx = log.SetLogger(ctx, logger)

	// update cluster meta info
	rkecluster.CreateLogPath = logPath
	rkecluster.PodCIDR = rkeConfig.Services.KubeController.ClusterCIDR
	rkecluster.ServiceCIDR = rkeConfig.Services.KubeController.ServiceClusterIPRange
	var kubernetesVersion = "v1.19.6-rke"
	if config.KubernetesVersion != "" {
		kubernetesVersion = config.KubernetesVersion
	}
	rkecluster.KubernetesVersion = kubernetesVersion
	nodeList, _ := json.Marshal(config.Nodes)
	rkecluster.NodeList = string(nodeList)
	rkecluster.NetworkMode = rkeConfig.Network.Plugin
	if err := r.Repo.Update(rkecluster); err != nil {
		logrus.Errorf("update rke cluster %s state failure %s", rkecluster.Name, err.Error())
	}

	// cluster init
	if err := cmd.ClusterInit(ctx, rkeConfig, hosts.DialersOptions{}, flags); err != nil {
		rollback("InitClusterConfig", err.Error(), "failure")
		rkecluster.Stats = v1alpha1.InstallFailed
		if err := r.Repo.Update(rkecluster); err != nil {
			logrus.Errorf("update rke cluster %s state failure %s", rkecluster.Name, err.Error())
		}
		return nil
	}
	rollback("InitClusterConfig", filePath, "success")

	// cluster install and up
	rollback("InstallKubernetes", filePath, "start")
	APIURL, _, _, _, configs, err := r.ClusterUp(ctx, hosts.DialersOptions{}, flags, map[string]interface{}{})
	if err != nil {
		rkecluster.Stats = v1alpha1.InstallFailed
		if err := r.Repo.Update(rkecluster); err != nil {
			logrus.Errorf("update rke cluster %s state failure %s", rkecluster.Name, err.Error())
		}
		rollback("InstallKubernetes", err.Error(), "failure")
		return nil
	}
	rkecluster.KubeConfig = configs[pki.KubeAdminCertName].Config
	rkecluster.APIURL = APIURL
	rkecluster.Stats = v1alpha1.RunningState
	rkecluster.NetworkMode = rkeConfig.Network.Plugin
	if err := r.Repo.Update(rkecluster); err != nil {
		logrus.Errorf("update rke cluster %s state failure %s", rkecluster.Name, err.Error())
	}
	rollback("InstallKubernetes", rkecluster.ClusterID, "success")
	return converClusterMeta(rkecluster)
}

func converClusterMeta(rkecluster *models.RKECluster) *v1alpha1.Cluster {
	var nodes v1alpha1.NodeList
	json.Unmarshal([]byte(rkecluster.NodeList), &nodes)
	cluster := &v1alpha1.Cluster{
		Name:              rkecluster.Name,
		ClusterID:         rkecluster.ClusterID,
		Created:           v1alpha1.NewTime(rkecluster.CreatedAt),
		MasterURL:         v1alpha1.MasterURL{APIServerEndpoint: rkecluster.APIURL},
		State:             rkecluster.Stats,
		ClusterType:       "rke",
		CurrentVersion:    rkecluster.KubernetesVersion,
		RegionID:          "",
		NetworkMode:       rkecluster.NetworkMode,
		SubnetCIDR:        rkecluster.ServiceCIDR,
		PodCIDR:           rkecluster.PodCIDR,
		KubernetesVersion: rkecluster.KubernetesVersion,
		CreateLogPath:     rkecluster.CreateLogPath,
		Size:              len(nodes),
		RainbondInit:      false,
		Parameters:        make(map[string]interface{}),
	}
	if rkecluster.KubeConfig != "" {
		kc := v1alpha1.KubeConfig{Config: rkecluster.KubeConfig}
		coreclient, _, err := kc.GetKubeClient()
		if err != nil {
			logrus.Errorf("create kube client failure %s", err.Error())
		}
		if coreclient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()
			versionByte, err := coreclient.RESTClient().Get().AbsPath("/version").DoRaw(ctx)
			var info version.Info
			json.Unmarshal(versionByte, &info)
			if err == nil {
				cluster.CurrentVersion = info.String()
			} else {
				cluster.State = v1alpha1.OfflineState
				cluster.Parameters["DisableRainbondInit"] = true
			}
			_, err = coreclient.CoreV1().ConfigMaps("rbd-system").Get(ctx, "region-config", metav1.GetOptions{})
			if err == nil {
				cluster.RainbondInit = true
			}
		}
	}
	return cluster
}

func (r *rkeAdaptor) CreateCluster(config v1alpha1.CreateClusterConfig) (*v1alpha1.Cluster, error) {
	rkeConfig, ok := config.(*v3.RancherKubernetesEngineConfig)
	if !ok {
		return nil, fmt.Errorf("cluster config is not RancherKubernetesEngineConfig")
	}

	var filePath string
	// setting up the flags
	flags := cluster.GetExternalFlags(false, false, false, false, "", filePath)
	// cluster init

	if err := cmd.ClusterInit(context.Background(), rkeConfig, hosts.DialersOptions{}, flags); err != nil {
		return nil, err
	}
	_, _, _, _, _, err := r.ClusterUp(context.Background(), hosts.DialersOptions{}, flags, map[string]interface{}{})
	return nil, err
}

func (r *rkeAdaptor) ClusterUp(ctx context.Context, dialersOptions hosts.DialersOptions, flags cluster.ExternalFlags, data map[string]interface{}) (string, string, string, string, map[string]pki.CertificatePKI, error) {
	var APIURL, caCrt, clientCert, clientKey string
	var reconcileCluster, restore bool

	clusterState, err := cluster.ReadStateFile(ctx, cluster.GetStateFilePath(flags.ClusterFilePath, flags.ConfigDir))
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	// We generate the first encryption config in ClusterInit, to store it ASAP. It's written to the DesiredState
	stateEncryptionConfig := clusterState.DesiredState.EncryptionConfig
	// if CurrentState has EncryptionConfig, it means this is NOT the first time we enable encryption, we should use the _latest_ applied value from the current cluster
	if clusterState.CurrentState.EncryptionConfig != "" {
		stateEncryptionConfig = clusterState.CurrentState.EncryptionConfig
	}

	kubeCluster, err := cluster.InitClusterObject(ctx, clusterState.DesiredState.RancherKubernetesEngineConfig.DeepCopy(), flags, stateEncryptionConfig)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	svcOptionsData := cluster.GetServiceOptionData(data)
	// check if rotate certificates is triggered
	if kubeCluster.RancherKubernetesEngineConfig.RotateCertificates != nil {
		return rebuildClusterWithRotatedCertificates(ctx, dialersOptions, flags, svcOptionsData)
	}
	// if we need to rotate the encryption key, do so and then return
	if kubeCluster.RancherKubernetesEngineConfig.RotateEncryptionKey {
		return RotateEncryptionKey(ctx, clusterState.CurrentState.RancherKubernetesEngineConfig.DeepCopy(), dialersOptions, flags)
	}

	log.Infof(ctx, "Building Kubernetes cluster")
	err = kubeCluster.SetupDialers(ctx, dialersOptions)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	err = kubeCluster.TunnelHosts(ctx, flags)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	currentCluster, err := kubeCluster.GetClusterState(ctx, clusterState)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	if !flags.DisablePortCheck {
		if err = kubeCluster.CheckClusterPorts(ctx, currentCluster); err != nil {
			return APIURL, caCrt, clientCert, clientKey, nil, err
		}
	}

	err = cluster.SetUpAuthentication(ctx, kubeCluster, currentCluster, clusterState)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	if len(kubeCluster.ControlPlaneHosts) > 0 {
		APIURL = fmt.Sprintf("https://%s:6443", kubeCluster.ControlPlaneHosts[0].Address)
	}
	clientCert = string(cert.EncodeCertPEM(kubeCluster.Certificates[pki.KubeAdminCertName].Certificate))
	clientKey = string(cert.EncodePrivateKeyPEM(kubeCluster.Certificates[pki.KubeAdminCertName].Key))
	caCrt = string(cert.EncodeCertPEM(kubeCluster.Certificates[pki.CACertName].Certificate))

	// moved deploying certs before reconcile to remove all unneeded certs generation from reconcile
	err = kubeCluster.SetUpHosts(ctx, flags)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	err = cluster.ReconcileCluster(ctx, kubeCluster, currentCluster, flags, svcOptionsData)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	/* reconcileCluster flag decides whether zero downtime upgrade logic is used or not.
	Zero-downtime upgrades should happen only when upgrading existing clusters. Not for new clusters or during etcd snapshot restore.
	currentCluster != nil indicates this is an existing cluster. Restore flag on DesiredState.RancherKubernetesEngineConfig indicates if it's a snapshot restore or not.
	reconcileCluster flag should be set to true only if currentCluster is not nil and restore is set to false
	*/
	if clusterState.DesiredState.RancherKubernetesEngineConfig != nil {
		restore = clusterState.DesiredState.RancherKubernetesEngineConfig.Restore.Restore
	}
	if currentCluster != nil && !restore {
		// reconcile this cluster, to check if upgrade is needed, or new nodes are getting added/removed
		/*This is to separate newly added nodes, so we don't try to check their status/cordon them before upgrade.
		This will also cover nodes that were considered inactive first time cluster was provisioned, but are now active during upgrade*/
		currentClusterNodes := make(map[string]bool)
		for _, node := range clusterState.CurrentState.RancherKubernetesEngineConfig.Nodes {
			currentClusterNodes[node.HostnameOverride] = true
		}

		newNodes := make(map[string]bool)
		for _, node := range clusterState.DesiredState.RancherKubernetesEngineConfig.Nodes {
			if !currentClusterNodes[node.HostnameOverride] {
				newNodes[node.HostnameOverride] = true
			}
		}
		kubeCluster.NewHosts = newNodes
		reconcileCluster = true

		maxUnavailableWorker, maxUnavailableControl, err := kubeCluster.CalculateMaxUnavailable()
		if err != nil {
			return APIURL, caCrt, clientCert, clientKey, nil, err
		}
		logrus.Infof("Setting maxUnavailable for worker nodes to: %v", maxUnavailableWorker)
		logrus.Infof("Setting maxUnavailable for controlplane nodes to: %v", maxUnavailableControl)
		kubeCluster.MaxUnavailableForWorkerNodes, kubeCluster.MaxUnavailableForControlNodes = maxUnavailableWorker, maxUnavailableControl
	}

	// update APIURL after reconcile
	if len(kubeCluster.ControlPlaneHosts) > 0 {
		APIURL = fmt.Sprintf("https://%s:6443", kubeCluster.ControlPlaneHosts[0].Address)
	}
	if err = cluster.ReconcileEncryptionProviderConfig(ctx, kubeCluster, currentCluster); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	if err := kubeCluster.PrePullK8sImages(ctx); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	errMsgMaxUnavailableNotFailedCtrl, err := kubeCluster.DeployControlPlane(ctx, svcOptionsData, reconcileCluster)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	// Apply Authz configuration after deploying controlplane
	err = cluster.ApplyAuthzResources(ctx, kubeCluster.RancherKubernetesEngineConfig, flags, dialersOptions)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	err = kubeCluster.UpdateClusterCurrentState(ctx, clusterState)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	err = cluster.SaveFullStateToKubernetes(ctx, kubeCluster, clusterState)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	errMsgMaxUnavailableNotFailedWrkr, err := kubeCluster.DeployWorkerPlane(ctx, svcOptionsData, reconcileCluster)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	if err = kubeCluster.CleanDeadLogs(ctx); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	err = kubeCluster.SyncLabelsAndTaints(ctx, currentCluster)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	err = cluster.ConfigureCluster(ctx, kubeCluster.RancherKubernetesEngineConfig, kubeCluster.Certificates, flags, dialersOptions, data, false)
	if err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}
	if kubeCluster.EncryptionConfig.RewriteSecrets {
		if err = kubeCluster.RewriteSecrets(ctx); err != nil {
			return APIURL, caCrt, clientCert, clientKey, nil, err
		}
	}

	if err := checkAllIncluded(kubeCluster); err != nil {
		return APIURL, caCrt, clientCert, clientKey, nil, err
	}

	if errMsgMaxUnavailableNotFailedCtrl != "" || errMsgMaxUnavailableNotFailedWrkr != "" {
		return APIURL, caCrt, clientCert, clientKey, nil, fmt.Errorf(errMsgMaxUnavailableNotFailedCtrl + errMsgMaxUnavailableNotFailedWrkr)
	}

	log.Infof(ctx, "Finished building Kubernetes cluster successfully")
	return APIURL, caCrt, clientCert, clientKey, kubeCluster.Certificates, nil
}

func checkAllIncluded(cluster *cluster.Cluster) error {
	if len(cluster.InactiveHosts) == 0 {
		return nil
	}

	var names []string
	for _, host := range cluster.InactiveHosts {
		names = append(names, host.Address)
	}

	if len(names) > 0 {
		return fmt.Errorf("Provisioning incomplete, host(s) [%s] skipped because they could not be contacted", strings.Join(names, ","))
	}
	return nil
}

func (r *rkeAdaptor) GetKubeConfig(clusterID string) (*v1alpha1.KubeConfig, error) {
	rkecluster, err := r.Repo.GetCluster(clusterID)
	if err != nil {
		return nil, fmt.Errorf("get cluster meta info failure %s", err.Error())
	}
	if rkecluster.KubeConfig == "" {
		return nil, fmt.Errorf("not found kube config")
	}
	return &v1alpha1.KubeConfig{Config: rkecluster.KubeConfig}, nil
}

func (r *rkeAdaptor) VPCList(regionID string) ([]*v1alpha1.VPC, error) {
	return nil, nil
}

func (r *rkeAdaptor) CreateVPC(v *v1alpha1.VPC) error {
	return nil
}

func (r *rkeAdaptor) DeleteVPC(regionID, vpcID string) error {
	return nil
}

func (r *rkeAdaptor) DescribeVPC(regionID, vpcID string) (*v1alpha1.VPC, error) {
	return nil, nil
}

func (r *rkeAdaptor) CreateVSwitch(v *v1alpha1.VSwitch) error {
	return nil
}

func (r *rkeAdaptor) DescribeVSwitch(regionID, vswitchID string) (*v1alpha1.VSwitch, error) {
	return nil, nil
}

func (r *rkeAdaptor) DeleteVSwitch(regionID, vswitchID string) error {
	return nil
}

func (r *rkeAdaptor) ListZones(regionID string) ([]*v1alpha1.Zone, error) {
	return nil, nil
}

func (r *rkeAdaptor) ListInstanceType(regionID string) ([]*v1alpha1.InstanceType, error) {
	return nil, nil
}

func (r *rkeAdaptor) CreateDB(db *v1alpha1.Database) error {
	return nil
}
