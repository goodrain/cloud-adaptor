package repo

import (
	"goodrain.com/cloud-adaptor/internal/model"
	"goodrain.com/cloud-adaptor/internal/repo/dao"
)

// AppStore
type AppStore struct {
	EID      string
	Name     string
	URL      string
	Branch   string
	Username string
	Password string
}

// AppStoreRepo -
type AppStoreRepo interface {
	Create(appStore *AppStore) error
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
		EID:      appStore.EID,
		Name:     appStore.Name,
		URL:      appStore.URL,
		Branch:   appStore.Branch,
		Username: appStore.Username,
		Password: appStore.Password,
	})
	if err != nil {
		return err
	}

	// TODO: create app templates
	return nil
}
