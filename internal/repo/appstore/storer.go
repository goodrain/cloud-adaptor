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
	"context"
	"sync"

	"golang.org/x/sync/singleflight"
	"goodrain.com/cloud-adaptor/internal/domain"
)

// Storer -
type Storer struct {
	singleflight.Group
	store        sync.Map
	appTemplater AppTemplater
}

// NewStorer creates a new Storer.
func NewStorer(appTemplater AppTemplater) *Storer {
	return &Storer{
		appTemplater: appTemplater,
	}
}

// Resync -
func (s *Storer) Resync(key string) {
	s.store.Delete(key)
}

// ListAppTemplates -
func (s *Storer) ListAppTemplates(ctx context.Context, appStore *domain.AppStore) ([]*domain.AppTemplate, error) {
	load, ok := s.store.Load(appStore.Key())
	if ok {
		appStore0, _ := load.(*domain.AppStore)
		if appStore0.Equals(appStore) {
			return appStore0.AppTemplates, nil
		}
	}

	appTemplates, err := s.appTemplater.Fetch(ctx, appStore)
	if err != nil {
		return nil, err
	}
	appStore.AppTemplates = appTemplates

	s.store.Store(appStore.Key(), appStore)

	return appTemplates, nil
}

func (s *Storer) DeleteAppTemplates(key string) {
	s.store.Delete(key)
}
