package dao

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"goodrain.com/cloud-adaptor/internal/model"
)

// AppStoreDao -
type AppStoreDao interface {
	Create(appStore *model.AppStore) error
	List(eid string) ([]*model.AppStore, error)
	Delete(name string) error
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

func (a *appStoreDao) List(eid string) ([]*model.AppStore, error) {
	var appStores []*model.AppStore
	if err := a.db.Where("eid=?", eid).Find(&appStores).Error; err != nil {
		return nil, errors.Wrap(err, "list app stores")
	}
	return appStores, nil
}
func (a *appStoreDao) Delete(name string) error {
	// TODO: handle 404 error
	return a.db.Where("name=?", name).Delete(&model.AppStore{}).Error
}
