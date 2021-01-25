package operator

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/goodrain/rainbond-operator/api/v1alpha1"
	rainbondv1alpha1 "github.com/goodrain/rainbond-operator/api/v1alpha1"
	"github.com/goodrain/rainbond-operator/util/commonutil"
	"github.com/goodrain/rainbond-operator/util/constants"
	"github.com/goodrain/rainbond-operator/util/rbdutil"
	"github.com/goodrain/rainbond-operator/util/retryutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//Operator operator
type Operator struct {
	Config
}

//Config operator config
type Config struct {
	RainbondVersion         string
	Namespace               string
	ArchiveFilePath         string
	RuntimeClient           client.Client
	Rainbondpackage         string
	RainbondImageRepository string
	ImageHubUser            string
	ImageHubPass            string
	OnlyInstallRegion       bool
}

//NewOperator new operator
func NewOperator(c Config) (*Operator, error) {
	if c.RuntimeClient == nil {
		return nil, fmt.Errorf("config runtime client can not be nil")
	}
	return &Operator{
		Config: c,
	}, nil
}

type componentClaim struct {
	namespace       string
	name            string
	version         string
	imageRepository string
	imageName       string
	Configs         map[string]string
	isInit          bool
	replicas        *int32
}

func (c *componentClaim) image() string {
	return path.Join(c.imageRepository, c.imageName) + ":" + c.version
}

func parseComponentClaim(claim *componentClaim) *v1alpha1.RbdComponent {
	component := &v1alpha1.RbdComponent{}
	component.Namespace = claim.namespace
	component.Name = claim.name
	component.Spec.Image = claim.image()
	component.Spec.ImagePullPolicy = corev1.PullIfNotPresent
	component.Spec.Replicas = claim.replicas
	labels := rbdutil.LabelsForRainbond(map[string]string{"name": claim.name})
	if claim.isInit {
		component.Spec.PriorityComponent = true
		labels["priorityComponent"] = "true"
	}
	component.Labels = labels
	return component
}

//Install install
func (o *Operator) Install(cluster *rainbondv1alpha1.RainbondCluster) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err := o.RuntimeClient.Create(ctx, cluster); err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("create rainbond cluster failure %s", err.Error())
		}
		var old rainbondv1alpha1.RainbondCluster
		if err := o.RuntimeClient.Get(ctx, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, &old); err != nil {
			return fmt.Errorf("get rainbond cluster failure %s", err.Error())
		}

		// Keep the image configuration
		if cluster.Spec.ImageHub == nil && old.Spec.ImageHub != nil {
			cluster.Spec.ImageHub = old.Spec.ImageHub
		}

		// Keep the database configuration
		if cluster.Spec.RegionDatabase == nil && old.Spec.RegionDatabase != nil {
			cluster.Spec.RegionDatabase = old.Spec.RegionDatabase
		}
		if cluster.Spec.UIDatabase == nil && old.Spec.UIDatabase != nil {
			cluster.Spec.UIDatabase = old.Spec.UIDatabase
		}
		old.Spec = cluster.Spec
		if err := o.RuntimeClient.Update(ctx, &old); err != nil {
			return fmt.Errorf("update rainbond cluster failure %s", err.Error())
		}
		*cluster = old
	}
	if err := o.createRainbondVolumes(cluster); err != nil {
		return fmt.Errorf("create rainbond volume failure %s", err.Error())
	}
	if err := o.createRainbondPackage(); err != nil {
		return fmt.Errorf("create rainbond volume failure %s", err.Error())
	}
	if err := o.createComponents(cluster); err != nil {
		return err
	}
	return nil
}

func (o *Operator) createComponents(cluster *v1alpha1.RainbondCluster) error {
	claims := o.genComponentClaims(cluster)
	for _, claim := range claims {
		// update image repository for priority components
		claim.imageRepository = cluster.Spec.RainbondImageRepository
		data := parseComponentClaim(claim)
		// init component
		data.Namespace = o.Namespace

		err := retryutil.Retry(time.Second*2, 3, func() (bool, error) {
			if err := o.createResourceIfNotExists(data); err != nil {
				return false, err
			}
			return true, nil
		})
		if err != nil {
			return fmt.Errorf("create rainbond component %s failure %s", data.GetName(), err.Error())
		}
	}
	return nil
}

func (o *Operator) genComponentClaims(cluster *v1alpha1.RainbondCluster) map[string]*componentClaim {
	var defReplicas = commonutil.Int32(1)
	if cluster.Spec.EnableHA {
		defReplicas = commonutil.Int32(2)
	}

	var isInit bool
	imageRepository := constants.DefImageRepository
	if cluster.Spec.ImageHub == nil || cluster.Spec.ImageHub.Domain == constants.DefImageRepository {
		isInit = true
	} else {
		imageRepository = path.Join(cluster.Spec.ImageHub.Domain, cluster.Spec.ImageHub.Namespace)
	}

	newClaim := func(name string) *componentClaim {
		defClaim := componentClaim{name: name, imageRepository: imageRepository, version: o.RainbondVersion, replicas: defReplicas}
		defClaim.imageName = name
		return &defClaim
	}
	name2Claim := map[string]*componentClaim{
		"rbd-api":            newClaim("rbd-api"),
		"rbd-chaos":          newClaim("rbd-chaos"),
		"rbd-eventlog":       newClaim("rbd-eventlog"),
		"rbd-monitor":        newClaim("rbd-monitor"),
		"rbd-mq":             newClaim("rbd-mq"),
		"rbd-worker":         newClaim("rbd-worker"),
		"rbd-webcli":         newClaim("rbd-webcli"),
		"rbd-resource-proxy": newClaim("rbd-resource-proxy"),
	}
	if !o.OnlyInstallRegion {
		name2Claim["rbd-app-ui"] = newClaim("rbd-app-ui")
	}
	name2Claim["metrics-server"] = newClaim("metrics-server")
	name2Claim["metrics-server"].version = "v0.3.6"

	if cluster.Spec.RegionDatabase == nil || (cluster.Spec.UIDatabase == nil && !o.OnlyInstallRegion) {
		claim := newClaim("rbd-db")
		claim.version = "8.0.19"
		claim.replicas = commonutil.Int32(1)
		name2Claim["rbd-db"] = claim
	}

	if cluster.Spec.ImageHub == nil || cluster.Spec.ImageHub.Domain == constants.DefImageRepository {
		claim := newClaim("rbd-hub")
		claim.imageName = "registry"
		claim.version = "2.6.2"
		claim.isInit = isInit
		name2Claim["rbd-hub"] = claim
	}

	name2Claim["rbd-gateway"] = newClaim("rbd-gateway")
	name2Claim["rbd-gateway"].isInit = isInit
	name2Claim["rbd-node"] = newClaim("rbd-node")
	name2Claim["rbd-node"].isInit = isInit

	if cluster.Spec.EtcdConfig == nil {
		claim := newClaim("rbd-etcd")
		claim.imageName = "etcd"
		claim.version = "v3.3.18"
		claim.isInit = isInit
		if cluster.Spec.EnableHA {
			claim.replicas = commonutil.Int32(3)
		}
		name2Claim["rbd-etcd"] = claim
	}

	// kubernetes dashboard
	k8sdashboard := newClaim("kubernetes-dashboard")
	k8sdashboard.version = "v2.0.1-3"
	name2Claim["kubernetes-dashboard"] = k8sdashboard
	dashboardscraper := newClaim("dashboard-metrics-scraper")
	dashboardscraper.imageName = "metrics-scraper"
	dashboardscraper.version = "v1.0.4"
	name2Claim["dashboard-metrics-scraper"] = dashboardscraper

	if rwx := cluster.Spec.RainbondVolumeSpecRWX; rwx != nil && rwx.CSIPlugin != nil {
		if rwx.CSIPlugin.NFS != nil {
			name2Claim["nfs-provisioner"] = newClaim("nfs-provisioner")
			name2Claim["nfs-provisioner"].replicas = commonutil.Int32(1)
			name2Claim["nfs-provisioner"].isInit = isInit
		}
		if rwx.CSIPlugin.AliyunNas != nil {
			name2Claim[constants.AliyunCSINasPlugin] = newClaim(constants.AliyunCSINasPlugin)
			name2Claim[constants.AliyunCSINasPlugin].isInit = isInit
			name2Claim[constants.AliyunCSINasProvisioner] = newClaim(constants.AliyunCSINasProvisioner)
			name2Claim[constants.AliyunCSINasProvisioner].isInit = isInit
			name2Claim[constants.AliyunCSINasProvisioner].replicas = commonutil.Int32(1)
		}
	}
	if rwo := cluster.Spec.RainbondVolumeSpecRWO; rwo != nil && rwo.CSIPlugin != nil {
		if rwo.CSIPlugin.AliyunCloudDisk != nil {
			name2Claim[constants.AliyunCSIDiskPlugin] = newClaim(constants.AliyunCSIDiskPlugin)
			name2Claim[constants.AliyunCSIDiskPlugin].isInit = isInit
			name2Claim[constants.AliyunCSIDiskProvisioner] = newClaim(constants.AliyunCSIDiskProvisioner)
			name2Claim[constants.AliyunCSIDiskProvisioner].isInit = isInit
			name2Claim[constants.AliyunCSIDiskProvisioner].replicas = commonutil.Int32(1)
		}
	}

	return name2Claim
}

func (o *Operator) createRainbondPackage() error {
	pkg := &v1alpha1.RainbondPackage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.Rainbondpackage,
			Namespace: o.Namespace,
		},
		Spec: v1alpha1.RainbondPackageSpec{
			PkgPath:      o.ArchiveFilePath,
			ImageHubUser: o.ImageHubUser,
			ImageHubPass: o.ImageHubPass,
		},
	}
	return o.createResourceIfNotExists(pkg)
}

func (o *Operator) createRainbondVolumes(cluster *v1alpha1.RainbondCluster) error {
	if cluster.Spec.RainbondVolumeSpecRWX != nil {
		rwx := setRainbondVolume("rainbondvolumerwx", o.Namespace, rbdutil.LabelsForAccessModeRWX(), cluster.Spec.RainbondVolumeSpecRWX)
		rwx.Spec.ImageRepository = o.RainbondImageRepository
		if err := o.createResourceIfNotExists(rwx); err != nil {
			return err
		}
	}
	if cluster.Spec.RainbondVolumeSpecRWO != nil {
		rwo := setRainbondVolume("rainbondvolumerwo", o.Namespace, rbdutil.LabelsForAccessModeRWO(), cluster.Spec.RainbondVolumeSpecRWO)
		rwo.Spec.ImageRepository = o.RainbondImageRepository
		if err := o.createResourceIfNotExists(rwo); err != nil {
			return err
		}
	}
	return nil
}

func (o *Operator) createResourceIfNotExists(resource client.Object) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := o.RuntimeClient.Create(ctx, resource)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		return fmt.Errorf("create resource %s/%s failure %s", resource.GetObjectKind(), resource.GetName(), err.Error())
	}
	return nil
}

func setRainbondVolume(name, namespace string, labels map[string]string, spec *v1alpha1.RainbondVolumeSpec) *v1alpha1.RainbondVolume {
	volume := &v1alpha1.RainbondVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    rbdutil.LabelsForRainbond(labels),
		},
		Spec: *spec,
	}
	return volume
}
