package domain

import (
	"github.com/helm/helm/pkg/repo"
	"github.com/pkg/errors"
	"goodrain.com/cloud-adaptor/pkg/bcode"
)

// AppStore
type AppStore struct {
	EID          string
	Name         string
	URL          string
	Branch       string
	Username     string
	Password     string
	AppTemplates []*AppTemplate
}

// Key -
func (a *AppStore) Key() string {
	return a.EID + a.Name
}

// Equals -
func (a *AppStore) Equals(b *AppStore) bool {
	if a.URL != b.URL {
		return false
	}
	if a.Branch != b.Branch {
		return false
	}
	if a.Username != b.Username {
		return false
	}
	if a.Password != b.Password {
		return false
	}
	return true
}

// GetAppTemplate get app template based on the app template name.
func (a *AppStore) GetAppTemplate(templateName string) (*AppTemplate, error) {
	for _, appTemplate := range a.AppTemplates {
		if appTemplate.Name == templateName {
			return appTemplate, nil
		}
	}
	return nil, errors.Wrap(bcode.ErrAppTemplateNotFound, "get app template")
}

// AppTemplate -
type AppTemplate struct {
	Name     string
	Versions []*repo.ChartVersion
}
