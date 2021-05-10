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
