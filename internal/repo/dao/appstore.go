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

package dao

import (
	"github.com/pkg/errors"
	"goodrain.com/cloud-adaptor/internal/model"
	"goodrain.com/cloud-adaptor/pkg/bcode"
	"gorm.io/gorm"
)

// AppStoreDao -
type AppStoreDao interface {
	Create(appStore *model.AppStore) error
	List(eid string) ([]*model.AppStore, error)
	Get(eid, name string) (*model.AppStore, error)
	Update(appStore *model.AppStore) error
	Delete(eid, name string) error
}

// NewAppStoreDao creates a new AppStoreDao
func NewAppStoreDao(db *gorm.DB) AppStoreDao {
	return &appStoreDao{
		db: db,
	}
}

type appStoreDao struct {
	db *gorm.DB
}

func (a *appStoreDao) Create(appStore *model.AppStore) error {
	err := a.db.Create(appStore).Error
	if err != nil {
		if isDuplicateEntry(err) {
			return errors.WithStack(bcode.ErrAppStoreNameConflict)
		}
		return errors.Wrap(err, "create app store")
	}
	return nil
}

func (a *appStoreDao) List(eid string) ([]*model.AppStore, error) {
	var appStores []*model.AppStore
	if err := a.db.Where("eid=?", eid).Find(&appStores).Error; err != nil {
		return nil, errors.Wrap(err, "list app stores")
	}
	return appStores, nil
}

func (a *appStoreDao) Get(eid, name string) (*model.AppStore, error) {
	var appStore model.AppStore
	if err := a.db.Where("eid=? and name=?", eid, name).Take(&appStore).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.WithStack(bcode.ErrAppStoreNotFound)
		}
		return nil, errors.Wrap(err, "get app store")
	}

	return &appStore, nil
}

func (a *appStoreDao) Update(appStore *model.AppStore) error {
	err := a.db.Save(appStore).Error
	if err != nil {
		if isDuplicateEntry(err) {
			return errors.WithStack(bcode.ErrAppStoreNameConflict)
		}
		return errors.Wrap(err, "update app store")
	}
	return nil
}

func (a *appStoreDao) Delete(eid, name string) error {
	err := a.db.Where("eid=? and name=?", eid, name).Delete(&model.AppStore{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.WithStack(bcode.ErrAppStoreNotFound)
		}
		return errors.Wrap(err, "delete app store")
	}
	return nil
}
