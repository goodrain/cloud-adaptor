package dao

import (
	"github.com/jinzhu/gorm"
	"goodrain.com/cloud-adaptor/internal/model"
)

// AppStoreDao -
type AppStoreDao interface {
	Create(appStore *model.AppStore) error
	foobar()
}

// NewAppStoreDao creates a new AppStoreDao
func NewAppStoreDao(db *gorm.DB) AppStoreDao {
	return &appStoreDao{
		db: db,
	}
}

type appStoreDao struct {
	db *gorm.DB
}

func (a *appStoreDao) Create(appStore *model.AppStore) error {
	// TODO: handle name conflict error
	return a.db.Create(appStore).Error
}

func (a *appStoreDao) foobar() {}
