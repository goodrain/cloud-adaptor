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
	"fmt"

	"goodrain.com/cloud-adaptor/internal/model"
	"goodrain.com/cloud-adaptor/pkg/util/uuidutil"
	"gorm.io/gorm"
)

// UpdateKubernetesTaskRepo enterprise create kubernetes task
type UpdateKubernetesTaskRepo struct {
	DB *gorm.DB `inject:""`
}

// NewUpdateKubernetesTaskRepo new Enterprise repoo
func NewUpdateKubernetesTaskRepo(db *gorm.DB) UpdateKubernetesTaskRepository {
	return &UpdateKubernetesTaskRepo{DB: db}
}

func (c *UpdateKubernetesTaskRepo) Transaction(tx *gorm.DB) UpdateKubernetesTaskRepository {
	return &UpdateKubernetesTaskRepo{DB: tx}
}

//Create create a task
func (c *UpdateKubernetesTaskRepo) Create(ck *model.UpdateKubernetesTask) error {
	var old model.UpdateKubernetesTask
	if ck.TaskID == "" {
		ck.TaskID = uuidutil.NewUUID()
	}
	if err := c.DB.Where("eid = ? and task_id=? and cluster_id=?", ck.EnterpriseID, ck.TaskID, ck.ClusterID).Take(&old).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// not found error, create new
			if err := c.DB.Save(ck).Error; err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return fmt.Errorf("task is exit")
}

//GetTaskByClusterID get cluster task
func (c *UpdateKubernetesTaskRepo) GetTaskByClusterID(eid string, providerName, clusterID string) (*model.UpdateKubernetesTask, error) {
	var old model.UpdateKubernetesTask
	if err := c.DB.Where("eid = ? and provider_name=? and cluster_id=?", eid, providerName, clusterID).Take(&old).Error; err != nil {
		return nil, err
	}
	return &old, nil
}

//UpdateStatus update status
func (c *UpdateKubernetesTaskRepo) UpdateStatus(eid string, taskID string, status string) error {
	var old model.UpdateKubernetesTask
	if err := c.DB.Model(&old).Where("eid = ? and task_id=?", eid, taskID).Update("status", status).Error; err != nil {
		return err
	}
	return nil
}

//GetTask get task
func (c *UpdateKubernetesTaskRepo) GetTask(eid string, taskID string) (*model.UpdateKubernetesTask, error) {
	var old model.UpdateKubernetesTask
	if err := c.DB.Where("eid = ? and task_id=?", eid, taskID).Take(&old).Error; err != nil {
		return nil, err
	}
	return &old, nil
}
