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

package usecase

import (
	"context"

	"goodrain.com/cloud-adaptor/internal/domain"
	"goodrain.com/cloud-adaptor/internal/repo"
)

// AppStoreUsecase -
type AppStoreUsecase struct {
	appStoreRepo repo.AppStoreRepo
}

// NewAppStoreUsecase -
func NewAppStoreUsecase(appStoreRepo repo.AppStoreRepo) *AppStoreUsecase {
	return &AppStoreUsecase{
		appStoreRepo: appStoreRepo,
	}
}

// Create -
func (a *AppStoreUsecase) Create(ctx context.Context, appStore *domain.AppStore) error {
	return a.appStoreRepo.Create(ctx, appStore)
}

// List -
func (a *AppStoreUsecase) List(ctx context.Context, eid string) ([]*domain.AppStore, error) {
	return a.appStoreRepo.List(eid)
}

// Update -
func (a *AppStoreUsecase) Update(ctx context.Context, appStore *domain.AppStore) error {
	return a.appStoreRepo.Update(ctx, appStore)
}

// Delete -
func (a *AppStoreUsecase) Delete(ctx context.Context, appStore *domain.AppStore) error {
	return a.appStoreRepo.Delete(appStore)
}

// Resync -
func (a *AppStoreUsecase) Resync(ctx context.Context, appStore *domain.AppStore) {
	a.appStoreRepo.Resync(appStore)
}

// GetAppTemplate returns the app template based on the app template name.
func (a *AppStoreUsecase) GetAppTemplate(ctx context.Context, appStore *domain.AppStore, appTemplateName string) (*domain.AppTemplate, error) {
	return appStore.GetAppTemplate(appTemplateName)
}
