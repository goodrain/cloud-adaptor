package dao

import (
	"github.com/pkg/errors"
	"goodrain.com/cloud-adaptor/internal/model"
	"goodrain.com/cloud-adaptor/pkg/bcode"
	"gorm.io/gorm"
)

// AppStoreDao -
type AppStoreDao interface {
	Create(appStore *model.AppStore) error
	List(eid string) ([]*model.AppStore, error)
	Get(eid, appStoreID string) (*model.AppStore, error)
	Update(appStore *model.AppStore) error
	Delete(eid, appStoreID string) error
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
	err := a.db.Create(appStore).Error
	if err != nil {
		if isDuplicateEntry(err) {
			return errors.WithStack(bcode.ErrAppStoreNameConflict)
		}
		return errors.Wrap(err, "create app store")
	}
	return nil
}

func (a *appStoreDao) List(eid string) ([]*model.AppStore, error) {
	var appStores []*model.AppStore
	if err := a.db.Where("eid=?", eid).Find(&appStores).Error; err != nil {
		return nil, errors.Wrap(err, "list app stores")
	}
	return appStores, nil
}

func (a *appStoreDao) Get(eid, appStoreID string) (*model.AppStore, error) {
	var appStore model.AppStore
	if err := a.db.Where("eid=? and appStoreID=?", eid, appStoreID).Take(&appStore).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.WithStack(bcode.ErrAppStoreNotFound)
		}
		return nil, errors.Wrap(err, "get app store")
	}

	return &appStore, nil
}

func (a *appStoreDao) Update(appStore *model.AppStore) error {
	err := a.db.Save(appStore).Error
	if err != nil {
		if isDuplicateEntry(err) {
			return errors.WithStack(bcode.ErrAppStoreNameConflict)
		}
		return errors.Wrap(err, "update app store")
	}
	return nil
}

func (a *appStoreDao) Delete(eid, appStoreID string) error {
	err := a.db.Where("eid=? and appStoreID=?", eid, appStoreID).Delete(&model.AppStore{}).Error
	if err != nil {
		return errors.Wrap(err, "delete app store")
	}
	return nil
}
