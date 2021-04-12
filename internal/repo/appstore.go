package repo

import (
	"context"

	"github.com/sirupsen/logrus"
	"goodrain.com/cloud-adaptor/internal/domain"
	"goodrain.com/cloud-adaptor/internal/model"
	"goodrain.com/cloud-adaptor/internal/repo/appstore"
	"goodrain.com/cloud-adaptor/internal/repo/dao"
	"goodrain.com/cloud-adaptor/pkg/util/uuidutil"
)

// AppStoreRepo -
type AppStoreRepo interface {
	Create(appStore *domain.AppStore) error
	List(eid string) ([]*domain.AppStore, error)
	Get(ctx context.Context, eid, appStoreID string) (*domain.AppStore, error)
	Delete(eid, appStoreID string) error
	Update(appStore *domain.AppStore) error
}

// NewAppStoreRepo creates a new AppStoreRepo.
func NewAppStoreRepo(appStoreDao dao.AppStoreDao, storer *appstore.Storer) AppStoreRepo {
	return &appStoreRepo{
		storer:      storer,
		appStoreDao: appStoreDao,
	}
}

type appStoreRepo struct {
	appStoreDao dao.AppStoreDao
	storer      *appstore.Storer
}

func (a *appStoreRepo) Create(appStore *domain.AppStore) error {
	// Create appStore
	err := a.appStoreDao.Create(&model.AppStore{
		EID:        appStore.EID,
		AppStoreID: uuidutil.NewUUID(),
		Name:       appStore.Name,
		URL:        appStore.URL,
		Branch:     appStore.Branch,
		Username:   appStore.Username,
		Password:   appStore.Password,
	})
	if err != nil {
		return err
	}

	// TODO: create app templates
	return nil
}

func (a *appStoreRepo) List(eid string) ([]*domain.AppStore, error) {
	appStores, err := a.appStoreDao.List(eid)
	if err != nil {
		return nil, err
	}

	var stores []*domain.AppStore
	for _, as := range appStores {
		stores = append(stores, &domain.AppStore{
			EID:        as.EID,
			AppStoreID: as.AppStoreID,
			Name:       as.Name,
			URL:        as.URL,
			Branch:     as.Branch,
			Username:   as.Username,
			Password:   as.Password,
		})
	}

	return stores, nil
}

func (a *appStoreRepo) Get(ctx context.Context, eid, appStoreID string) (*domain.AppStore, error) {
	as, err := a.appStoreDao.Get(eid, appStoreID)
	if err != nil {
		return nil, err
	}

	// TODO: deduplicate the code below
	appStore := &domain.AppStore{
		EID:        as.EID,
		AppStoreID: as.AppStoreID,
		Name:       as.Name,
		URL:        as.URL,
		Branch:     as.Branch,
		Username:   as.Username,
		Password:   as.Password,
	}

	appStore.AppTemplates, err = a.storer.ListAppTemplates(ctx, appStore)
	if err != nil {
		logrus.Warningf("[appStoreRepo] [Get] list app templates: %v", err)
	}

	return appStore, nil
}

func (a *appStoreRepo) Update(appStore *domain.AppStore) error {
	as, err := a.appStoreDao.Get(appStore.EID, appStore.AppStoreID)
	if err != nil {
		return err
	}
	as.Name = appStore.Name
	as.URL = appStore.URL
	as.Branch = appStore.Branch
	as.Username = appStore.Username
	as.Password = appStore.Password

	return a.appStoreDao.Update(as)
}
func (a *appStoreRepo) Delete(eid, appStoreID string) error {
	return a.appStoreDao.Delete(eid, appStoreID)
}
