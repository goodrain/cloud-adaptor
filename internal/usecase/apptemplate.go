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
