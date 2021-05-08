package usecase

import (
	"context"
	"goodrain.com/cloud-adaptor/internal/domain"
	"goodrain.com/cloud-adaptor/internal/repo"
)

// AppTemplate -
type AppTemplate struct {
	templateVersionRepo repo.TemplateVersionRepo
}

// NewAppTemplate -
func NewAppTemplate(templateVersionRepo repo.TemplateVersionRepo) *AppTemplate {
	return &AppTemplate{
		templateVersionRepo: templateVersionRepo,
	}
}

// GetVersion returns the app template version.
func (a *AppTemplate) GetVersion(ctx context.Context, appStore *domain.AppStore, templateName, version string) (*domain.AppTemplateVersion, error) {
	return a.templateVersionRepo.GetTemplateVersion(appStore.Name, appStore.URL, templateName, version)
}
