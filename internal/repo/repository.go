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

package repo

import (
	"goodrain.com/cloud-adaptor/internal/model"
	"gorm.io/gorm"
)

//CloudAccesskeyRepository enterprise accesskey repository
type CloudAccesskeyRepository interface {
	Create(ent *model.CloudAccessKey) error
	GetByProviderAndEnterprise(providerName, eid string) (*model.CloudAccessKey, error)
}

//CreateKubernetesTaskRepository enterprise create kubernetes task
type CreateKubernetesTaskRepository interface {
	Transaction(tx *gorm.DB) CreateKubernetesTaskRepository
	Create(ent *model.CreateKubernetesTask) error
	GetLastTask(eid string, providerName string) (*model.CreateKubernetesTask, error)
	UpdateStatus(eid string, taskID string, status string) error
	GetTask(eid string, taskID string) (*model.CreateKubernetesTask, error)
	GetLatestOneByName(name string) (*model.CreateKubernetesTask, error)
}

//InitRainbondTaskRepository init rainbond region task
type InitRainbondTaskRepository interface {
	Transaction(tx *gorm.DB) InitRainbondTaskRepository
	Create(ent *model.InitRainbondTask) error
	GetTaskByClusterID(eid string, providerName, clusterID string) (*model.InitRainbondTask, error)
	UpdateStatus(eid string, taskID string, status string) error
	GetTask(eid string, taskID string) (*model.InitRainbondTask, error)
	DeleteTask(eid string, providerName, clusterID string) error
	GetTaskRunningLists(eid string) ([]*model.InitRainbondTask, error)
}

//UpdateKubernetesTaskRepository -
type UpdateKubernetesTaskRepository interface {
	Transaction(tx *gorm.DB) UpdateKubernetesTaskRepository
	Create(ent *model.UpdateKubernetesTask) error
	GetTaskByClusterID(eid string, providerName, clusterID string) (*model.UpdateKubernetesTask, error)
	UpdateStatus(eid string, taskID string, status string) error
	GetTask(eid string, taskID string) (*model.UpdateKubernetesTask, error)
}

//TaskEventRepository task event
type TaskEventRepository interface {
	Transaction(tx *gorm.DB) TaskEventRepository
	Create(ent *model.TaskEvent) error
	ListEvent(eid, taskID string) ([]*model.TaskEvent, error)
}

//RainbondClusterConfigRepository -
type RainbondClusterConfigRepository interface {
	Create(ent *model.RainbondClusterConfig) error
	Get(clusterID string) (*model.RainbondClusterConfig, error)
}

// RKEClusterRepository -
type RKEClusterRepository interface {
	Create(te *model.RKECluster) error
	Update(te *model.RKECluster) error
	GetCluster(eid, name string) (*model.RKECluster, error)
	ListCluster(eid string) ([]*model.RKECluster, error)
	DeleteCluster(eid, name string) error
}
