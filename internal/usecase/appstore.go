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
func (a *AppStoreUsecase) Delete(ctx context.Context, eid, name string) error {
	return a.appStoreRepo.Delete(eid, name)
}

// Resync -
func (a *AppStoreUsecase) Resync(ctx context.Context, appStore *domain.AppStore) {
	a.appStoreRepo.Resync(appStore)
}

// GetAppTemplate returns the app template based on the app template name.
func (a *AppStoreUsecase) GetAppTemplate(ctx context.Context, appStore *domain.AppStore, appTemplateName string) (*domain.AppTemplate, error) {
	return appStore.GetAppTemplate(appTemplateName)
}
