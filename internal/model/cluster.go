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

package model

//RKECluster RKE cluster
type RKECluster struct {
	Model
	EnterpriseID      string `gorm:"column:eid" json:"eid"`
	Name              string `gorm:"column:name" json:"name,omitempty"`
	ClusterID         string `gorm:"column:clusterID" json:"clusterID,omitempty"`
	APIURL            string `gorm:"column:apiURL;type:text" json:"apiURL,omitempty"`
	KubeConfig        string `gorm:"column:kubeConfig;type:text" json:"kubeConfig,omitempty"`
	NetworkMode       string `gorm:"column:networkMode" json:"networkMode,omitempty"`
	ServiceCIDR       string `gorm:"column:serviceCIDR" json:"serviceCIDR,omitempty"`
	PodCIDR           string `gorm:"column:podCIDR" json:"podCIDR,omitempty"`
	KubernetesVersion string `gorm:"column:kubernetesVersion" json:"kubernetesVersion,omitempty"`
	RainbondInit      bool   `gorm:"column:rainbondInit" json:"rainbondInit,omitempty"`
	CreateLogPath     string `gorm:"column:createLogPath" json:"createLogPath,omitempty"`
	// Deprecated
	NodeList          string `gorm:"column:nodeList;type:text" json:"nodeList,omitempty"`
	Stats             string `gorm:"column:stats" json:"stats,omitempty"`
	RKEConfig         string `gorm:"column:rkeConfig"`
}

//CustomCluster custom cluster
type CustomCluster struct {
	Model
	EnterpriseID string `gorm:"column:eid" json:"eid"`
	Name         string `gorm:"column:name" json:"name,omitempty"`
	ClusterID    string `gorm:"column:clusterID" json:"clusterID,omitempty"`
	KubeConfig   string `gorm:"column:kubeConfig;type:text" json:"kubeConfig,omitempty"`
	EIP          string `gorm:"column:eip" json:"eip,omitempty"`
}

//RainbondClusterConfig rainbond cluster config
type RainbondClusterConfig struct {
	Model
	EnterpriseID string `gorm:"column:eid" json:"eid"`
	ClusterID    string `gorm:"column:clusterID" json:"clusterID,omitempty"`
	Config       string `gorm:"column:config;type:text" json:"config,omitempty"`
}
