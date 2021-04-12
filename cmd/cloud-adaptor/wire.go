// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"goodrain.com/cloud-adaptor/internal/biz"
	"goodrain.com/cloud-adaptor/internal/handler"
	"goodrain.com/cloud-adaptor/internal/middleware"
	"goodrain.com/cloud-adaptor/internal/nsqc"
	"goodrain.com/cloud-adaptor/internal/repo"
	"goodrain.com/cloud-adaptor/internal/repo/dao"
	"goodrain.com/cloud-adaptor/internal/task"
	"goodrain.com/cloud-adaptor/internal/types"
	"gorm.io/gorm"
)

// initApp init the application.
func initApp(context.Context,
	*gorm.DB,
	chan types.KubernetesConfigMessage,
	chan types.InitRainbondConfigMessage,
	chan types.UpdateKubernetesConfigMessage) (*gin.Engine, error) {
	panic(wire.Build(handler.ProviderSet, biz.ProviderSet, repo.ProviderSet, task.ProviderSet,
		nsqc.ProviderSet, dao.ProviderSet, middleware.ProviderSet, newApp))
}
