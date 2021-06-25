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

package repo

import (
	"context"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"goodrain.com/cloud-adaptor/cmd/cloud-adaptor/config"
	"goodrain.com/cloud-adaptor/internal/domain"
	"goodrain.com/cloud-adaptor/internal/model"
	"goodrain.com/cloud-adaptor/internal/repo/appstore"
	"goodrain.com/cloud-adaptor/internal/repo/dao"
	"goodrain.com/cloud-adaptor/pkg/bcode"
)

// AppStoreRepo -
type AppStoreRepo interface {
	Create(ctx context.Context, appStore *domain.AppStore) error
	List(eid string) ([]*domain.AppStore, error)
	Get(ctx context.Context, eid, name string) (*domain.AppStore, error)
	Delete(appStore *domain.AppStore) error
	Update(ctx context.Context, appStore *domain.AppStore) error
	Resync(appStore *domain.AppStore)
}

// NewAppStoreRepo creates a new AppStoreRepo.
func NewAppStoreRepo(cfg *config.Config, appStoreDao dao.AppStoreDao, storer *appstore.Storer, appTemplater appstore.AppTemplater) AppStoreRepo {
	return &appStoreRepo{
		cfg:          cfg,
		storer:       storer,
		appStoreDao:  appStoreDao,
		appTemplater: appTemplater,
	}
}

type appStoreRepo struct {
	cfg          *config.Config
	appStoreDao  dao.AppStoreDao
	storer       *appstore.Storer
	appTemplater appstore.AppTemplater
}

func (a *appStoreRepo) Create(ctx context.Context, appStore *domain.AppStore) error {
	// Check the availability of the app store.
	if err := a.isAvailable(ctx, appStore); err != nil {
		return err
	}

	return a.appStoreDao.Create(&model.AppStore{
		EID:      appStore.EID,
		Name:     appStore.Name,
		URL:      appStore.URL,
		Branch:   appStore.Branch,
		Username: appStore.Username,
		Password: appStore.Password,
	})
}

func (a *appStoreRepo) List(eid string) ([]*domain.AppStore, error) {
	appStores, err := a.appStoreDao.List(eid)
	if err != nil {
		return nil, err
	}

	var stores []*domain.AppStore
	for _, as := range appStores {
		stores = append(stores, &domain.AppStore{
			EID:      as.EID,
			Name:     as.Name,
			URL:      as.URL,
			Branch:   as.Branch,
			Username: as.Username,
			Password: as.Password,
		})
	}

	return stores, nil
}

func (a *appStoreRepo) Get(ctx context.Context, eid, name string) (*domain.AppStore, error) {
	as, err := a.appStoreDao.Get(eid, name)
	if err != nil {
		return nil, err
	}

	// TODO: deduplicate the code below
	appStore := &domain.AppStore{
		EID:      as.EID,
		Name:     as.Name,
		URL:      as.URL,
		Branch:   as.Branch,
		Username: as.Username,
		Password: as.Password,
	}

	appStore.AppTemplates, err = a.storer.ListAppTemplates(ctx, appStore)
	if err != nil {
		logrus.Warningf("[appStoreRepo] [Get] list app templates: %v", err)
	}

	return appStore, nil
}

func (a *appStoreRepo) Update(ctx context.Context, appStore *domain.AppStore) error {
	if err := a.isAvailable(ctx, appStore); err != nil {
		return err
	}

	as, err := a.appStoreDao.Get(appStore.EID, appStore.Name)
	if err != nil {
		return err
	}
	as.Name = appStore.Name
	as.URL = appStore.URL
	as.Branch = appStore.Branch
	as.Username = appStore.Username
	as.Password = appStore.Password

	return a.appStoreDao.Update(as)
}

func (a *appStoreRepo) Delete(appStore *domain.AppStore) error {
	// delete from database
	if err := a.appStoreDao.Delete(appStore.EID, appStore.Name); err != nil && !errors.Is(err, bcode.ErrAppStoreNotFound) {
		return err
	}

	// delete from cache
	a.storer.DeleteAppTemplates(appStore.Key())

	// delete charts in cache
	dest := path.Join(a.cfg.Helm.RepoCache, appStore.Name)
	if err := os.RemoveAll(dest); err != nil {
		logrus.Warningf("key: %s; delete charts cache: %v", appStore.Key(), err)
	}

	return nil
}

func (a *appStoreRepo) Resync(appStore *domain.AppStore) {
	a.storer.Resync(appStore.Key())
}

func (a *appStoreRepo) isAvailable(ctx context.Context, appStore *domain.AppStore) error {
	_, err := a.appTemplater.Fetch(ctx, appStore)
	if err != nil {
		return errors.Wrap(bcode.ErrAppStoreUnavailable, err.Error())
	}
	return nil
}
