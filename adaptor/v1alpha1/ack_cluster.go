package v1alpha1

import (
	"net"
	"strings"
)

//GetDefaultACKCreateClusterConfig get create ack cluster default config
func GetDefaultACKCreateClusterConfig(config KubernetesClusterConfig) CreateClusterConfig {
	kubernetesVersion := "1.16.9-aliyun.1"
	if config.KubernetesVersion != "" && strings.Contains(config.KubernetesVersion, "aliyun") {
		kubernetesVersion = config.KubernetesVersion
	}
	dockerVersion := "19.03.5"
	if config.DockerVersion != "" {
		dockerVersion = config.DockerVersion
	}
	serviceClusterIPRange := "172.21.0.0/20"
	podIPRange := "172.20.0.0/16"
	if config.ServiceCIDR != "" {
		if _, _, err := net.ParseCIDR(config.ServiceCIDR); err == nil {
			serviceClusterIPRange = config.ServiceCIDR
		}
	}
	if config.ClusterCIDR != "" {
		if _, _, err := net.ParseCIDR(config.ClusterCIDR); err == nil {
			podIPRange = config.ClusterCIDR
		}
	}
	return &AckClusterConfig{
		Name:                 config.ClusterName,
		DisableRollback:      true,
		ClusterType:          ManagedKubernetes,
		TimeoutMins:          60,
		KubernetesVersion:    kubernetesVersion,
		RegionID:             config.Region,
		SNATEntry:            true,
		CloudMonitorFlags:    true,
		EndpointPublicAccess: true,
		DeletionProtection:   true,
		NodeCidrMask:         "25",
		ProxyMode:            "ipvs",
		Tags:                 []string{},
		Addons: []Addon{
			{
				Name: "flananel",
			},
			{
				Name: "csi-plugin",
			},
			{
				Name: "csi-provisioner",
			},
			{
				Name:    "nginx-ingress-controller",
				Disable: true,
			},
		},
		OSType:   "Linux",
		Platform: "CentOS",
		Runtime: Runtime{
			Name:    "docker",
			Version: dockerVersion,
		},
		WorkerInstanceType: []string{config.InstanceType},
		NumOfNodes: func() int {
			if config.WorkerNodeNum < 2 {
				return 2
			}
			return config.WorkerNodeNum
		}(),
		WorkerSystemDiskCategory: "cloud_efficiency",
		WorkerSystemDiskSize:     120,
		WorkerDataDisks: []WorkerDataDisk{
			{
				Category:  "cloud_efficiency",
				Size:      "200",
				Encrypted: "false",
			},
		},
		WorkerInstanceChargeType: "PostPaid",
		ContainerCIDR:            podIPRange,
		ServiceCIDR:              serviceClusterIPRange,
		CPUPolicy:                "none",
		VPCID:                    config.VpcID,
		VSwitchIDs:               []string{config.VSwitchID},
		LoginPassword:            "RootPassword123!",
	}
}
