package repo

import (
	"goodrain.com/cloud-adaptor/internal/model"
	"goodrain.com/cloud-adaptor/internal/repo/dao"
	"goodrain.com/cloud-adaptor/pkg/util/uuidutil"
)

// AppStore
type AppStore struct {
	EID        string
	AppStoreID string
	Name       string
	URL        string
	Branch     string
	Username   string
	Password   string
}

// AppStoreRepo -
type AppStoreRepo interface {
	Create(appStore *AppStore) error
	List(eid string) ([]*AppStore, error)
	Get(eid, appStoreID string) (*AppStore, error)
	Delete(eid, appStoreID string) error
}

// NewAppStoreRepo creates a new AppStoreRepo.
func NewAppStoreRepo(appStoreDao dao.AppStoreDao) AppStoreRepo {
	return &appStoreRepo{
		appStoreDao: appStoreDao,
	}
}

type appStoreRepo struct {
	appStoreDao dao.AppStoreDao
}

func (a *appStoreRepo) Create(appStore *AppStore) error {
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

func (a *appStoreRepo) List(eid string) ([]*AppStore, error) {
	appStores, err := a.appStoreDao.List(eid)
	if err != nil {
		return nil, err
	}

	var stores []*AppStore
	for _, as := range appStores {
		stores = append(stores, &AppStore{
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

func (a *appStoreRepo) Get(eid, appStoreID string) (*AppStore, error) {
	as, err := a.appStoreDao.Get(eid, appStoreID)
	if err != nil {
		return nil, err
	}

	// TODO: deduplicate the code below
	return &AppStore{
		EID:      as.EID,
		Name:     as.Name,
		URL:      as.URL,
		Branch:   as.Branch,
		Username: as.Username,
		Password: as.Password,
	}, nil
}

func (a *appStoreRepo) Delete(eid, appStoreID string) error {
	return a.appStoreDao.Delete(eid, appStoreID)
}
