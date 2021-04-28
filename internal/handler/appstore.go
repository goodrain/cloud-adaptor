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
// @Summary creates a new app store.
// @Tags appstores
// @ID createAppStore
// @Accept  json
// @Produce  json
// @Param eid path string true "the enterprise id"
// @Param createAppStoreReq body v1.CreateAppStoreReq true "."
// @Success 200 {object} v1.AppStore
// @Failure 400 {object} ginutil.Result "8002, app store unavailable"
// @Failure 409 {object} ginutil.Result "8001, app store name conflict"
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/enterprises/:eid/appstores [post]
func (a *AppStoreHandler) Create(c *gin.Context) {
	var req v1.CreateAppStoreReq
	if err := ginutil.ShouldBindJSON(c, &req); err != nil {
		ginutil.Error(c, err)
		return
	}

	eid := c.Param("eid")

	// DTO to DO
	// TODO: Code generation or reflection
	appStore := &domain.AppStore{
		EID:      eid,
		Name:     req.Name,
		URL:      req.URL,
		Branch:   req.Branch,
		Username: req.Username,
		Password: req.Password,
	}
	err := a.appStore.Create(c.Request.Context(), appStore)

	ginutil.JSON(c, &v1.AppStore{
		EID:      appStore.EID,
		Name:     appStore.Name,
		URL:      appStore.URL,
		Branch:   appStore.Branch,
		Username: appStore.Username,
		Password: appStore.Password,
	}, err)
}

// List returns a list of app stores.
// @Summary returns a list of app stores.
// @Tags appstores
// @ID listAppStores
// @Accept  json
// @Produce  json
// @Param eid path string true "the enterprise id"
// @Success 200 {array} v1.AppStore
// @Router /api/v1/enterprises/:eid/appstores [get]
func (a *AppStoreHandler) List(c *gin.Context) {
	eid := c.Param("eid")
	appStores, err := a.appStore.List(c.Request.Context(), eid)

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

	ginutil.JSON(c, stores, err)
}

// Get returns the app store.
// @Summary returns the app store.
// @Tags appstores
// @ID getAppStore
// @Accept  json
// @Produce  json
// @Param eid path string true "the enterprise id"
// @Param name path string true "the name of the app store"
// @Success 200 {object} v1.AppStore
// @Failure 404 {object} ginutil.Result "8000, app store not found"
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/enterprises/:eid/appstores/:name [get]
func (a *AppStoreHandler) Get(c *gin.Context) {
	appStore := ginutil.MustGet(c, "appStore").(*domain.AppStore)
	ginutil.JSON(c, appStore)
}

// Update updates the app store.
// @Summary updates the app store..
// @Tags appstores
// @ID updateAppStore
// @Accept  json
// @Produce  json
// @Param eid path string true "the enterprise id"
// @Param name path string true "the name of the app store"
// @Param updateAppStoreReq body v1.UpdateAppStoreReq true "."
// @Success 200 {object} v1.AppStore
// @Failure 404 {object} ginutil.Result "8000, app store not found"
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/enterprises/:eid/appstores/:name [put]
func (a *AppStoreHandler) Update(c *gin.Context) {
	var req v1.UpdateAppStoreReq
	if err := ginutil.ShouldBindJSON(c, &req); err != nil {
		ginutil.Error(c, err)
		return
	}

	appStoreI, _ := c.Get("appStore")
	appStore := appStoreI.(*domain.AppStore)

	appStore.URL = req.URL
	appStore.Branch = req.Branch
	appStore.Username = req.Username
	appStore.Password = req.Password

	ginutil.JSON(c, nil, a.appStore.Update(c.Request.Context(), appStore))
}

// Delete deletes the app store.
// @Summary deletes the app store.
// @Tags appstores
// @ID deleteAppStore
// @Accept  json
// @Produce  json
// @Param eid path string true "the enterprise id"
// @Param name path string true "the name of the app store"
// @Success 200
// @Failure 404 {object} ginutil.Result "8000, app store not found"
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/enterprises/:eid/appstores/:name [delete]
func (a *AppStoreHandler) Delete(c *gin.Context) {
	appStore := ginutil.MustGet(c, "appStore").(*domain.AppStore)
	ginutil.JSON(c, nil, a.appStore.Delete(c.Request.Context(), appStore.EID, appStore.Name))
}

// Resync resync the app templates.
// @Summary resync the app templates.
// @Tags appstores
// @ID resyncAppStore
// @Accept  json
// @Produce  json
// @Param eid path string true "the enterprise id"
// @Param name path string true "the name of the app store"
// @Success 200
// @Failure 404 {object} ginutil.Result "8000, app store not found"
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/enterprises/:eid/appstores/:name/resync [post]
func (a *AppStoreHandler) Resync(c *gin.Context) {
	appStore := ginutil.MustGet(c, "appStore").(*domain.AppStore)
	a.appStore.Resync(c.Request.Context(), appStore)
}

// ListTemplates returns a list of app templates.
// @Summary returns a list of app templates.
// @Tags appstores
// @ID listTemplates
// @Accept  json
// @Produce  json
// @Param eid path string true "the enterprise id"
// @Param name path string true "the name of the app store"
// @Success 200 {array} v1.AppTemplate
// @Failure 404 {object} ginutil.Result "8000, app store not found"
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/enterprises/:eid/appstores/:name/apps [get]
func (a *AppStoreHandler) ListTemplates(c *gin.Context) {
	appStore := ginutil.MustGet(c, "appStore").(*domain.AppStore)

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

// GetAppTemplate returns the app template.
// @Summary returns the app template.
// @Tags appstores
// @ID getAppTemplate
// @Accept  json
// @Produce  json
// @Param eid path string true "the enterprise id"
// @Param name path string true "the name of the app store"
// @Param templateName path string true "the name of the app template"
// @Success 200 {object} v1.AppTemplate
// @Failure 404 {object} ginutil.Result "8000, app store not found"
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/enterprises/:eid/appstores/:name/apps/:templateName [get]
func (a *AppStoreHandler) GetAppTemplate(c *gin.Context) {
	appStore := ginutil.MustGet(c, "appStore").(*domain.AppStore)
	appTemplate, err := a.appStore.GetAppTemplate(c.Request.Context(), appStore, c.Param("templateName"))
	if err != nil {
		ginutil.Error(c, err)
		return
	}

	ginutil.JSON(c, &v1.AppTemplate{
		Name:     appTemplate.Name,
		Versions: appTemplate.Versions,
	})
}
