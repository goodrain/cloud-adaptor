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

package rke

import (
	"fmt"

	"goodrain.com/cloud-adaptor/internal/model"
	"goodrain.com/cloud-adaptor/pkg/util/uuidutil"
	"gorm.io/gorm"
)

// ClusterRepo -
type ClusterRepo struct {
	DB *gorm.DB `inject:""`
}

// NewRKEClusterRepo new Enterprise repoo
func NewRKEClusterRepo(db *gorm.DB) *ClusterRepo {
	return &ClusterRepo{DB: db}
}

//Create create an event
func (t *ClusterRepo) Create(te *model.RKECluster) error {
	var old model.RKECluster
	if err := t.DB.Where("name=?", te.Name).Find(&old).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// not found error, create new
			if te.ClusterID == "" {
				te.ClusterID = uuidutil.NewUUID()
			}
			if err := t.DB.Save(te).Error; err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return fmt.Errorf("rke cluster %s is exist", te.Name)
}

//Update -
func (t *ClusterRepo) Update(te *model.RKECluster) error {
	return t.DB.Save(te).Error
}

//GetCluster -
func (t *ClusterRepo) GetCluster(name string) (*model.RKECluster, error) {
	var rc model.RKECluster
	if err := t.DB.Where("name=? or clusterID=?", name, name).Find(&rc).Error; err != nil {
		return nil, err
	}
	return &rc, nil
}

//ListCluster -
func (t *ClusterRepo) ListCluster() ([]*model.RKECluster, error) {
	var list []*model.RKECluster
	if err := t.DB.Order("created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

//DeleteCluster delete cluster
func (t *ClusterRepo) DeleteCluster(name string) error {
	var rc model.RKECluster
	if err := t.DB.Where("name=? or clusterID=?", name, name).Delete(&rc).Error; err != nil {
		return err
	}
	return nil
}
