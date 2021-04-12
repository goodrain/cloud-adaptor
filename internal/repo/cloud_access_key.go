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
	"gorm.io/gorm"
	"goodrain.com/cloud-adaptor/internal/model"
)

// CloudAccessKeyRepo enterprise cloud accesskey repo
type CloudAccessKeyRepo struct {
	DB *gorm.DB
}

// NewCloudAccessKeyRepo new Enterprise repoo
func NewCloudAccessKeyRepo(db *gorm.DB) CloudAccesskeyRepository {
	return &CloudAccessKeyRepo{DB: db}
}

//Create create, Keep an enterprise with the same provider have one accesskey
func (c *CloudAccessKeyRepo) Create(ck *model.CloudAccessKey) error {
	var old model.CloudAccessKey
	if err := c.DB.Where("eid = ? and provider_name=?", ck.EnterpriseID, ck.ProviderName).Find(&old).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// not found error, create new
			if err := c.DB.Save(ck).Error; err != nil {
				return err
			}
			return nil
		}
		return err
	}
	old.AccessKey = ck.AccessKey
	old.SecretKey = ck.SecretKey
	*ck = old
	return c.DB.Save(&old).Error
}

//GetByProviderAndEnterprise get
func (c *CloudAccessKeyRepo) GetByProviderAndEnterprise(providerName, eid string) (*model.CloudAccessKey, error) {
	var old model.CloudAccessKey
	if err := c.DB.Where("eid = ? and provider_name=?", eid, providerName).Find(&old).Error; err != nil {
		return nil, err
	}
	return &old, nil
}
