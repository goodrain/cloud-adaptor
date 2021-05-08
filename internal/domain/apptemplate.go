package domain

import (
	"github.com/helm/helm/pkg/repo"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
)

// AppTemplateVersion is a domain object of app template version.
type AppTemplateVersion struct {
	repo.ChartVersion
	Readme    string
	Questions []v3.Question
	Values    map[string]string
}

// AppTemplate -
type AppTemplate struct {
	Name     string
	Versions []*repo.ChartVersion
}
