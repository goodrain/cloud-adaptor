package biz

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

func (a *AppStoreUsecase) Create(ctx context.Context, appStore *domain.AppStore) error {
	return a.appStoreRepo.Create(ctx, appStore)
}

func (a *AppStoreUsecase) List(eid string) ([]*domain.AppStore, error) {
	return a.appStoreRepo.List(eid)
}

func (a *AppStoreUsecase) Update(ctx context.Context, appStore *domain.AppStore) error {
	return a.appStoreRepo.Update(ctx, appStore)
}

func (a *AppStoreUsecase) Delete(eid, name string) error {
	return a.appStoreRepo.Delete(eid, name)
}

func (a *AppStoreUsecase) Resync(appStore *domain.AppStore) {
	a.appStoreRepo.Resync(appStore)
}
