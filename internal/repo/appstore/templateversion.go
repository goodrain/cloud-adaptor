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

package appstore

import (
	"github.com/goodrain/rainbond/pkg/helm"
	"github.com/pkg/errors"
	"goodrain.com/cloud-adaptor/cmd/cloud-adaptor/config"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// TemplateVersioner -
type TemplateVersioner struct {
	repoFile  string
	repoCache string
}

// NewTemplateVersioner creates a new TemplateVersioner.
func NewTemplateVersioner(cfg *config.Config) *TemplateVersioner {
	return &TemplateVersioner{
		repoFile:  cfg.Helm.RepoFile,
		repoCache: cfg.Helm.RepoCache,
	}
}

// LoadChart loads chart.
func (t *TemplateVersioner) LoadChart(repo, repoURL, templateName, version string) (*chart.Chart, error) {
	helmCmd, err := helm.NewHelm("nonamespace", t.repoFile, t.repoCache)
	if err != nil {
		return nil, errors.WithMessage(err, "create a new helm command")
	}

	repoCmd := helm.NewRepo(t.repoFile, t.repoCache)
	if err := repoCmd.Add(repo, repoURL, "", ""); err != nil {
		return nil, errors.WithMessage(err, "add helm repo")
	}

	cp, err := helmCmd.Load(repo+"/"+templateName, version)
	if err != nil {
		return nil, err
	}

	return loader.Load(cp)
}
