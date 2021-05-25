// RAINBOND, Application Management Platform
// Copyright (C) 2020-2020 Goodrain Co., Ltd.

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

package cluster

import (
	rainbondv1alpha1 "github.com/goodrain/rainbond-operator/api/v1alpha1"
	"goodrain.com/cloud-adaptor/adaptor/v1alpha1"
	"goodrain.com/cloud-adaptor/api/models"
	v1 "goodrain.com/cloud-adaptor/api/openapi/types/v1"
)

//Usecase represents the cluster's usecases
type Usecase interface {
	ListKubernetesCluster(string, v1.ListKubernetesCluster) ([]*v1alpha1.Cluster, error)
	GetCluster(provider, eid, clusterID string) (*v1alpha1.Cluster, error)
	CreateKubernetesCluster(eid string, req v1.CreateKubernetesReq) (*models.CreateKubernetesTask, error)
	InstallCluster(eid, clusterID string) (*models.CreateKubernetesTask, error)
	AddAccessKey(eid string, key v1.AddAccessKey) (*models.CloudAccessKey, error)
	GetByProviderAndEnterprise(providerName, eid string) (*models.CloudAccessKey, error)
	CreateTaskEvent(em *v1.EventMessage) (*models.TaskEvent, error)
	ListTaskEvent(eid, taskID string) ([]*models.TaskEvent, error)
	GetLastCreateKubernetesTask(eid, providerName string) (*models.CreateKubernetesTask, error)
	GetCreateKubernetesTask(eid, taskID string) (*models.CreateKubernetesTask, error)
	InitRainbondRegion(eid string, req v1.InitRainbondRegionReq) (*models.InitRainbondTask, error)
	GetTaskRunningLists(eid string) ([]*models.InitRainbondTask, error)
	GetInitRainbondTaskByClusterID(eid, clusterID, providerName string) (*models.InitRainbondTask, error)
	GetRegionConfig(eid, clusterID, providerName string) (map[string]string, error)
	UpdateInitRainbondTaskStatus(eid, taskID, status string) (*models.InitRainbondTask, error)
	DeleteKubernetesCluster(eid, clusterID, provider string) error
	GetKubeConfig(eid, clusterID, providerName string) (string, error)
	UpdateKubernetesCluster(eid string, req v1.UpdateKubernetesReq) (*models.UpdateKubernetesTask, error)
	GetUpdateKubernetesTaskByClusterID(eid, clusterID, providerName string) (*models.UpdateKubernetesTask, error)
	GetRKENodeList(eid, clusterID string) (v1alpha1.NodeList, error)
	SetRainbondClusterConfig(eid, clusterID, config string) error
	GetRainbondClusterConfig(eid, clusterID string) (*rainbondv1alpha1.RainbondCluster, string)
	UninstallRainbondRegion(eid, clusterID, provider string) error
}
