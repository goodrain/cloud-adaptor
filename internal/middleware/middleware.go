package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"goodrain.com/cloud-adaptor/internal/repo"
	"goodrain.com/cloud-adaptor/pkg/util/ginutil"
)

// ProviderSet is a middleware provider.
var ProviderSet = wire.NewSet(NewMiddleware)

// Middleware -
type Middleware struct {
	appStoreRepo repo.AppStoreRepo
}

// NewAppStore creates a new middleware.
func NewMiddleware(appStoreRepo repo.AppStoreRepo) *Middleware {
	return &Middleware{
		appStoreRepo: appStoreRepo,
	}
}

func (a *Middleware) AppStore(c *gin.Context) {
	eid := c.Param("eid")
	appStoreID := c.Param("appStoreID")
	appStore, err := a.appStoreRepo.Get(eid, appStoreID)
	if err != nil {
		ginutil.Error(c, err)
		return
	}
	c.Set("appStore", appStore)
}
