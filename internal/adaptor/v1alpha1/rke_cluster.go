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

package v1alpha1

import (
	"fmt"
	"net"

	v3 "github.com/rancher/rke/types"
	"goodrain.com/cloud-adaptor/version"
)

//GetDefaultRKECreateClusterConfig get default rke create cluster config
func GetDefaultRKECreateClusterConfig(config KubernetesClusterConfig) CreateClusterConfig {
	var nodeMaps = make(map[string]v3.RKEConfigNode, len(config.Nodes))
	for _, node := range config.Nodes {
		nodeMaps[node.IP] = v3.RKEConfigNode{
			NodeName: "",
			Address:  node.IP,
			Port: func() string {
				if node.SSHPort != 0 {
					return fmt.Sprintf("%d", node.SSHPort)
				}
				return "22"
			}(),
			DockerSocket: node.DockerSocketPath,
			User: func() string {
				if node.SSHUser != "" {
					return node.SSHUser
				}
				return "docker"
			}(),
			SSHKeyPath:      "~/.ssh/id_rsa",
			Role:            node.Roles,
			InternalAddress: node.InternalAddress,
		}
	}
	serviceClusterIPRange := "10.43.0.0/16"
	podIPRange := "10.42.0.0/16"
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
	var kubernetesVersion = "v1.23.10-rancher1"
	if config.KubernetesVersion != "" {
		kubernetesVersion = config.KubernetesVersion
	}
	var networkMode = "flannel"
	if config.NetworkMode != "" && (config.NetworkMode == "flannel" || config.NetworkMode == "calico") {
		networkMode = config.NetworkMode
	}
	var rkeConfig = &v3.RancherKubernetesEngineConfig{
		//default 45, Depending on network and node configuration factors, the startup time may be long.
		//so, We need to expand the timeout to 5 minutes
		AddonJobTimeout: 60 * 5,
		ClusterName:     config.ClusterName,
		Nodes: func() []v3.RKEConfigNode {
			var nodes []v3.RKEConfigNode
			for k := range nodeMaps {
				nodes = append(nodes, nodeMaps[k])
			}
			return nodes
		}(),
		Services: v3.RKEConfigServices{
			Etcd: v3.ETCDService{
				BaseService: v3.BaseService{
					ExtraEnv: []string{"ETCD_AUTO_COMPACTION_RETENTION=1"},
				},
			},
			KubeAPI: v3.KubeAPIService{ServiceClusterIPRange: serviceClusterIPRange},
			// KubeController Service
			KubeController: v3.KubeControllerService{ClusterCIDR: podIPRange, ServiceClusterIPRange: serviceClusterIPRange},
			// Scheduler Service
			Scheduler: v3.SchedulerService{},
			// Kubelet Service
			Kubelet: v3.KubeletService{
				BaseService: v3.BaseService{
					ExtraBinds: []string{"/grlocaldata:/grlocaldata:rw,z", "/cache:/cache:rw,z"},
				},
				ClusterDomain:    "cluster.local",
				ClusterDNSServer: "10.43.0.10",
			},
			// KubeProxy Service
			Kubeproxy: v3.KubeproxyService{},
		},
		Network: v3.NetworkConfig{
			Plugin: networkMode,
		},
		Authentication: v3.AuthnConfig{
			Strategy: "x509",
		},
		SystemImages: v3.RKESystemImages{
			// etcd image
			Etcd: version.InstallImageRepo + "/mirrored-coreos-etcd:v3.5.3",
			// Alpine image
			Alpine: version.InstallImageRepo + "/rke-tools:v0.1.87",
			// rke-nginx-proxy image
			NginxProxy: version.InstallImageRepo + "/rke-tools:v0.1.87",
			// rke-cert-deployer image
			CertDownloader: version.InstallImageRepo + "/rke-tools:v0.1.87",
			// rke-service-sidekick image
			KubernetesServicesSidecar: version.InstallImageRepo + "/rke-tools:v0.1.87",
			// KubeDNS image
			KubeDNS: "rancher/mirrored-k8s-dns-kube-dns:1.21.1",
			// DNSMasq image
			DNSmasq: "rancher/mirrored-k8s-dns-dnsmasq-nanny:1.21.1",
			// KubeDNS side car image
			KubeDNSSidecar: "rancher/mirrored-k8s-dns-sidecar:1.21.1",
			// KubeDNS autoscaler image
			KubeDNSAutoscaler: version.InstallImageRepo + "/mirrored-cluster-proportional-autoscaler:1.8.5",
			// CoreDNS image
			CoreDNS: version.InstallImageRepo + "/mirrored-coredns-coredns:1.9.0",
			// CoreDNS autoscaler image
			CoreDNSAutoscaler: version.InstallImageRepo + "/mirrored-cluster-proportional-autoscaler:1.8.5",
			// Nodelocal image
			Nodelocal: version.InstallImageRepo + "/mirrored-k8s-dns-node-cache:1.21.1",
			// Kubernetes image
			Kubernetes: version.InstallImageRepo + "/hyperkube:" + kubernetesVersion,
			// Flannel image
			Flannel: version.InstallImageRepo + "/mirrored-coreos-flannel:v0.15.1",
			// Flannel CNI image
			FlannelCNI: version.InstallImageRepo + "/flannel-cni:v0.3.0-rancher6",
			// Calico Node image
			CalicoNode: "",
			// Calico CNI image
			CalicoCNI: "",
			// Calico Controllers image
			CalicoControllers: "",
			// Calicoctl image
			CalicoCtl: "",
			//CalicoFlexVol image
			CalicoFlexVol: "",
			// Canal Node Image
			CanalNode: "rancher/mirrored-calico-node:v3.22.0",
			// Canal CNI image
			CanalCNI: "",
			// Canal Controllers Image needed for Calico/Canal v3.14.0+
			CanalControllers: "",
			//CanalFlannel image
			CanalFlannel: "",
			//CanalFlexVol image
			CanalFlexVol: "",
			//Weave Node image
			WeaveNode: "",
			// Weave CNI image
			WeaveCNI: "",
			// Pod infra container image
			PodInfraContainer: version.InstallImageRepo + "/mirrored-pause:3.6",
			// Ingress Controller image
			Ingress: "",
			// Ingress Controller Backend image
			IngressBackend: "",
			// Metrics Server image
			MetricsServer: version.InstallImageRepo + "/mirrored-metrics-server:v0.6.1",
			// Pod infra container image for Windows
			WindowsPodInfraContainer: "",
			// Cni deployer container image for Cisco ACI
			AciCniDeployContainer: "",
			// host container image for Cisco ACI
			AciHostContainer: "",
			// opflex agent container image for Cisco ACI
			AciOpflexContainer: "",
			// mcast daemon container image for Cisco ACI
			AciMcastContainer: "",
			// OpenvSwitch container image for Cisco ACI
			AciOpenvSwitchContainer: "",
			// Controller container image for Cisco ACI
			AciControllerContainer: "",
			// GBP Server container image for Cisco ACI
			AciGbpServerContainer: "",
			// Opflex Server container image for Cisco ACI
			AciOpflexServerContainer: "",
		},
		Authorization: v3.AuthzConfig{Mode: "rbac"},
		Ingress: v3.IngressConfig{
			Provider: "none",
		},
		Monitoring: v3.MonitoringConfig{
			Provider: "none",
		},
	}
	return rkeConfig
}
