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

package repository

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"goodrain.com/cloud-adaptor/api/cluster"
	"goodrain.com/cloud-adaptor/api/models"
	"goodrain.com/cloud-adaptor/util/uuidutil"
)

// InitRainbondRegionTaskRepo enterprise create kubernetes task
type InitRainbondRegionTaskRepo struct {
	DB *gorm.DB `inject:""`
}

// NewInitRainbondRegionTaskRepo new Enterprise repoo
func NewInitRainbondRegionTaskRepo(db *gorm.DB) cluster.InitRainbondTaskRepository {
	return &InitRainbondRegionTaskRepo{DB: db}
}

//Create create a task
func (c *InitRainbondRegionTaskRepo) Create(ck *models.InitRainbondTask) error {
	var old models.InitRainbondTask
	if ck.TaskID == "" {
		ck.TaskID = uuidutil.NewUUID()
	}
	if err := c.DB.Where("eid = ? and task_id=? and cluster_id=?", ck.EnterpriseID, ck.TaskID, ck.ClusterID).Find(&old).Error; err != nil {
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
func (c *InitRainbondRegionTaskRepo) GetTaskByClusterID(eid string, providerName, clusterID string) (*models.InitRainbondTask, error) {
	var old models.InitRainbondTask
	if err := c.DB.Where("eid = ? and provider_name=? and cluster_id=?", eid, providerName, clusterID).Find(&old).Error; err != nil {
		return nil, err
	}
	return &old, nil
}

//UpdateStatus update status
func (c *InitRainbondRegionTaskRepo) UpdateStatus(eid string, taskID string, status string) error {
	var old models.InitRainbondTask
	if err := c.DB.Model(&old).Where("eid = ? and task_id=?", eid, taskID).Update("status", status).Error; err != nil {
		return err
	}
	return nil
}

//GetTask get task
func (c *InitRainbondRegionTaskRepo) GetTask(eid string, taskID string) (*models.InitRainbondTask, error) {
	var old models.InitRainbondTask
	if err := c.DB.Where("eid = ? and task_id=?", eid, taskID).Find(&old).Error; err != nil {
		return nil, err
	}
	return &old, nil
}

//GetTaskRunningLists get not complete tasks
func (c *InitRainbondRegionTaskRepo) GetTaskRunningLists(eid string) ([]*models.InitRainbondTask, error) {
	var list []*models.InitRainbondTask
	if err := c.DB.Where("eid = ? and status != ?", eid, "complete").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

//DeleteTask -
func (c *InitRainbondRegionTaskRepo) DeleteTask(eid string, providerName, clusterID string) error {
	var old models.InitRainbondTask
	if err := c.DB.Where("eid = ? and provider_name=? and cluster_id=?", eid, providerName, clusterID).Delete(&old).Error; err != nil {
		return err
	}
	return nil
}
