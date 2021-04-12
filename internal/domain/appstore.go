package domain

import "github.com/helm/helm/pkg/repo"

// AppStore
type AppStore struct {
	EID          string
	AppStoreID   string
	Name         string
	URL          string
	Branch       string
	Username     string
	Password     string
	AppTemplates []*AppTemplate
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

// AppTemplate -
type AppTemplate struct {
	Name     string
	Versions []*repo.ChartVersion
}
