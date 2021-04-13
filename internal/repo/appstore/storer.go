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
	store sync.Map
}

// NewStorer creates a new Storer.
func NewStorer() *Storer {
	return &Storer{}
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

	appTemplater := NewAppTemplater(ctx, appStore)
	// single flight to avoid cache breakdown
	v, err, _ := s.Do(appStore.AppStoreID, func() (interface{}, error) {
		return appTemplater.Fetch()
	})
	if err != nil {
		return nil, err
	}
	appTemplates := v.([]*domain.AppTemplate)
	appStore.AppTemplates = appTemplates

	s.store.Store(appStore.Key(), appStore)

	return appTemplates, nil
}
