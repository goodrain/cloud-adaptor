package handler

import (
	"github.com/gin-gonic/gin"
	v1 "goodrain.com/cloud-adaptor/api/cloud-adaptor/v1"
	"goodrain.com/cloud-adaptor/internal/biz"
	"goodrain.com/cloud-adaptor/internal/repo"
	"goodrain.com/cloud-adaptor/pkg/bcode"
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
func (a *AppStoreHandler) Create(ctx *gin.Context) {
	var req v1.CreateAppStoreReq
	// TODO: Wrap in ginutil, return bcode directly
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}

	eid := ctx.Param("eid")

	// DTO to DO
	appStore := &repo.AppStore{
		EID:      eid,
		Name:     req.Name,
		URL:      req.URL,
		Branch:   req.Branch,
		Username: req.Username,
		Password: req.Password,
	}
	err := a.appStore.Create(appStore)

	// TODO: use ginutil.JSONv2
	ginutil.JSON(ctx, &v1.CreateAppStoreResp{
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

}

// Delete deletes the app store.
func (a *AppStoreHandler) Delete(ctx *gin.Context) {
	name := ctx.Param("name")
	err := a.appStore.Delete(name)
	ginutil.JSON(ctx, nil, err)
}
