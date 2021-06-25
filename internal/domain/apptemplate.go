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
