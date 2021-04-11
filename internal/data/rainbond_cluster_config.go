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

package data

import (
	"github.com/jinzhu/gorm"
	"goodrain.com/cloud-adaptor/internal/model"
)

// RainbondClusterConfigRepo enterprise task event
type RainbondClusterConfigRepo struct {
	DB *gorm.DB `inject:""`
}

// NewRainbondClusterConfigRepo new Enterprise repoo
func NewRainbondClusterConfigRepo(db *gorm.DB) RainbondClusterConfigRepository {
	return &RainbondClusterConfigRepo{DB: db}
}

//Create create an event
func (t *RainbondClusterConfigRepo) Create(te *model.RainbondClusterConfig) error {
	var old model.RainbondClusterConfig
	if err := t.DB.Where("clusterID=?", te.ClusterID).Find(&old).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := t.DB.Save(te).Error; err != nil {
				return err
			}
			return nil
		}
		return err
	}
	old.Config = te.Config
	return t.DB.Save(old).Error
}

//Get -
func (t *RainbondClusterConfigRepo) Get(clusterID string) (*model.RainbondClusterConfig, error) {
	var rcc model.RainbondClusterConfig
	if err := t.DB.Where("clusterID=?", clusterID).Find(&rcc).Error; err != nil {
		return nil, err
	}
	return &rcc, nil
}
