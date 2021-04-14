package handler

import (
	"github.com/gin-gonic/gin"
	v1 "goodrain.com/cloud-adaptor/api/cloud-adaptor/v1"
	"goodrain.com/cloud-adaptor/internal/biz"
	"goodrain.com/cloud-adaptor/internal/domain"
	"goodrain.com/cloud-adaptor/pkg/util/ginutil"
)

// AppStoreHandler -
type AppStoreHandler struct {
	appStore *biz.AppStoreUsecase
}

// NewClusterHandler new enterprise handler
func NewAppStoreHandler(appStore *biz.AppStoreUsecase) *AppStoreHandler {
	return &AppStoreHandler{
		appStore: appStore,
	}
}

// Create creates a new app store.
func (a *AppStoreHandler) Create(c *gin.Context) {
	var req v1.CreateAppStoreReq
	if err := ginutil.ShouldBindJSON(c, &req); err != nil {
		ginutil.Error(c, err)
		return
	}

	eid := c.Param("eid")

	// DTO to DO
	appStore := &domain.AppStore{
		EID:      eid,
		Name:     req.Name,
		URL:      req.URL,
		Branch:   req.Branch,
		Username: req.Username,
		Password: req.Password,
	}
	err := a.appStore.Create(appStore)

	ginutil.JSON(c, &v1.AppStore{
		EID:      appStore.EID,
		Name:     appStore.Name,
		URL:      appStore.URL,
		Branch:   appStore.Branch,
		Username: appStore.Username,
		Password: appStore.Password,
	}, err)
}

// Create creates a new app store.
func (a *AppStoreHandler) List(ctx *gin.Context) {
	eid := ctx.Param("eid")
	appStores, err := a.appStore.List(eid)

	var stores []*v1.AppStore
	for _, as := range appStores {
		stores = append(stores, &v1.AppStore{
			EID:      as.EID,
			Name:     as.Name,
			URL:      as.URL,
			Branch:   as.Branch,
			Username: as.Username,
			Password: as.Password,
		})
	}

	ginutil.JSON(ctx, stores, err)
}

// Get deletes the app store.
func (a *AppStoreHandler) Get(c *gin.Context) {
	appStore := ginutil.MustGet(c, "appStore").(*domain.AppStore)
	ginutil.JSON(c, appStore)
}

// Update updates the app store.
func (a *AppStoreHandler) Update(c *gin.Context) {
	var req v1.UpdateAppStoreReq
	if err := ginutil.ShouldBindJSON(c, &req); err != nil {
		ginutil.Error(c, err)
		return
	}

	appStoreI, _ := c.Get("appStore")
	appStore := appStoreI.(*domain.AppStore)

	appStore.Name = req.Name
	appStore.URL = req.URL
	appStore.Branch = req.Branch
	appStore.Username = req.Username
	appStore.Password = req.Password

	ginutil.JSON(c, nil, a.appStore.Update(appStore))
}

// Delete deletes the app store.
func (a *AppStoreHandler) Delete(c *gin.Context) {
	appStore := ginutil.MustGet(c, "appStore").(*domain.AppStore)
	ginutil.JSON(c, nil, a.appStore.Delete(appStore.EID, appStore.Name))
}

func (a *AppStoreHandler) Resync(c *gin.Context) {
	appStore := ginutil.MustGet(c, "appStore").(*domain.AppStore)
	a.appStore.Resync(appStore)
}

func (a *AppStoreHandler) ListTemplates(c *gin.Context) {
	appStoreI, _ := c.Get("appStore")
	appStore := appStoreI.(*domain.AppStore)

	appTemplates := appStore.AppTemplates
	var templates []*v1.AppTemplate
	for _, at := range appTemplates {
		templates = append(templates, &v1.AppTemplate{
			Name:     at.Name,
			Versions: at.Versions,
		})
	}

	ginutil.JSON(c, templates)
}
