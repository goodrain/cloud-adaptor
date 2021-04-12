package biz

import (
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

func (a *AppStoreUsecase) Create(appStore *repo.AppStore) error {
	return a.appStoreRepo.Create(appStore)
}

func (a *AppStoreUsecase) List(eid string) ([]*repo.AppStore, error) {
	return a.appStoreRepo.List(eid)
}

func (a *AppStoreUsecase) Delete(name string) error {
	return a.appStoreRepo.Delete(name)
}
