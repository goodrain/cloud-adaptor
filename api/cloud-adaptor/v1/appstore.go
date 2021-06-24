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

package v1

import (
	"github.com/helm/helm/pkg/repo"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
)

// CreateAppStoreReq -
type CreateAppStoreReq struct {
	// The name of app store.
	Name string `json:"name" binding:"required,appStoreName"`
	// The url of app store.
	URL string `json:"url" binding:"required"`
	// The branch of app store, which category is git repo.
	Branch string `json:"branch"`
	// The username of the private app store
	Username string `json:"username"`
	// The password of the private app store
	Password string `json:"password"`
}

// UpdateAppStoreReq -
type UpdateAppStoreReq struct {
	// The url of app store.
	URL string `json:"url" binding:"required"`
	// The branch of app store, which category is git repo.
	Branch string `json:"branch"`
	// The username of the private app store
	Username string `json:"username"`
	// The password of the private app store
	Password string `json:"password"`
}

// AppStore -
type AppStore struct {
	// The enterprise id.
	EID string `json:"eid"`
	// The name of app store.
	Name string `json:"name"`
	// The url of app store.
	URL string `json:"url"`
	// The branch of app store, which category is git repo.
	Branch string `json:"branch"`
	// The username of the private app store
	Username string `json:"username"`
	// The password of the private app store
	Password string `json:"password"`
}

// AppTemplate -
type AppTemplate struct {
	// The name of app template.
	Name string `json:"name"`
	// A list of app template versions.
	Versions []*repo.ChartVersion `json:"versions"`
}

// TemplateVersion represents a app template version.
type TemplateVersion struct {
	repo.ChartVersion
	// The readme content of the chart.
	Readme string `json:"readme"`
	// The questions content of the chart
	Questions []v3.Question `json:"questions"`
	// A list of values files.
	Values map[string]string `json:"values"`
}
