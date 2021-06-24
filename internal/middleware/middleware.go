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

package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"goodrain.com/cloud-adaptor/internal/repo"
	"goodrain.com/cloud-adaptor/pkg/util/ginutil"
)

// ProviderSet is a middleware provider.
var ProviderSet = wire.NewSet(NewMiddleware)

// Middleware -
type Middleware struct {
	appStoreRepo repo.AppStoreRepo
}

// NewMiddleware creates a new middleware.
func NewMiddleware(appStoreRepo repo.AppStoreRepo) *Middleware {
	return &Middleware{
		appStoreRepo: appStoreRepo,
	}
}

// AppStore -
func (a *Middleware) AppStore(c *gin.Context) {
	eid := c.Param("eid")
	name := c.Param("name")
	appStore, err := a.appStoreRepo.Get(c.Request.Context(), eid, name)
	if err != nil {
		ginutil.Error(c, err)
		return
	}
	c.Set("appStore", appStore)
}
