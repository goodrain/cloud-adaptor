// RAINBOND, Application Management Platform
// Copyright (C) 2014-2017 Goodrain Co., Ltd.

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

package adaptor

import (
	"context"
	"fmt"

	rainbondv1alpha1 "github.com/goodrain/rainbond-operator/api/v1alpha1"
	"goodrain.com/cloud-adaptor/adaptor/v1alpha1"
)

var (
	//ErrNotSupportRDS not support rds
	ErrNotSupportRDS = fmt.Errorf("not support rds")
)

//CloudAdaptor cloud adaptor interface
type CloudAdaptor interface {
	RainbondClusterAdaptor
	VPCList(regionID string) ([]*v1alpha1.VPC, error)
	CreateVPC(v *v1alpha1.VPC) error
	DeleteVPC(regionID, vpcID string) error
	DescribeVPC(regionID, vpcID string) (*v1alpha1.VPC, error)
	CreateVSwitch(v *v1alpha1.VSwitch) error
	DescribeVSwitch(regionID, vswitchID string) (*v1alpha1.VSwitch, error)
	DeleteVSwitch(regionID, vswitchID string) error
	ListZones(regionID string) ([]*v1alpha1.Zone, error)
	ListInstanceType(regionID string) ([]*v1alpha1.InstanceType, error)
	CreateDB(*v1alpha1.Database) error
}

//KubernetesClusterAdaptor -
type KubernetesClusterAdaptor interface {
	ClusterList() ([]*v1alpha1.Cluster, error)
	DescribeCluster(clusterID string) (*v1alpha1.Cluster, error)
	CreateCluster(v1alpha1.CreateClusterConfig) (*v1alpha1.Cluster, error)
	GetKubeConfig(clusterID string) (*v1alpha1.KubeConfig, error)
	DeleteCluster(clusterID string) error
}

//RainbondClusterAdaptor rainbond init adaptor
type RainbondClusterAdaptor interface {
	KubernetesClusterAdaptor
	CreateRainbondKubernetes(ctx context.Context, config *v1alpha1.KubernetesClusterConfig, rollback func(step, message, status string)) *v1alpha1.Cluster
	GetRainbondInitConfig(cluster *v1alpha1.Cluster, gateway, chaos []*rainbondv1alpha1.K8sNode, rollback func(step, message, status string)) *v1alpha1.RainbondInitConfig
}
