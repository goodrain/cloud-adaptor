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
	"goodrain.com/cloud-adaptor/api/models"
)

//CloudAccesskeyRepository enterprise accesskey repository
type CloudAccesskeyRepository interface {
	Create(ent *models.CloudAccessKey) error
	GetByProviderAndEnterprise(providerName, eid string) (*models.CloudAccessKey, error)
}

//CreateKubernetesTaskRepository enterprise create kubernetes task
type CreateKubernetesTaskRepository interface {
	Create(ent *models.CreateKubernetesTask) error
	GetLastTask(eid string, providerName string) (*models.CreateKubernetesTask, error)
	UpdateStatus(eid string, taskID string, status string) error
	GetTask(eid string, taskID string) (*models.CreateKubernetesTask, error)
}

//InitRainbondTaskRepository init rainbond region task
type InitRainbondTaskRepository interface {
	Create(ent *models.InitRainbondTask) error
	GetTaskByClusterID(eid string, providerName, clusterID string) (*models.InitRainbondTask, error)
	UpdateStatus(eid string, taskID string, status string) error
	GetTask(eid string, taskID string) (*models.InitRainbondTask, error)
	GetTaskRunningLists(eid string) ([]*models.InitRainbondTask, error)
}

//UpdateKubernetesTaskRepository -
type UpdateKubernetesTaskRepository interface {
	Create(ent *models.UpdateKubernetesTask) error
	GetTaskByClusterID(eid string, providerName, clusterID string) (*models.UpdateKubernetesTask, error)
	UpdateStatus(eid string, taskID string, status string) error
	GetTask(eid string, taskID string) (*models.UpdateKubernetesTask, error)
}

//TaskEventRepository task event
type TaskEventRepository interface {
	Create(ent *models.TaskEvent) error
	ListEvent(eid, taskID string) ([]*models.TaskEvent, error)
}
