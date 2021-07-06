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
	"net"
	"os"
	"strings"
)

//GetDefaultACKCreateClusterConfig get create ack cluster default config
func GetDefaultACKCreateClusterConfig(config KubernetesClusterConfig) CreateClusterConfig {
	defaultAckVersion := os.Getenv("DEFAULT_ACK_VERSION")
	if defaultAckVersion == "" {
		defaultAckVersion = "1.18.8-aliyun.1"
	}
	kubernetesVersion := defaultAckVersion
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
