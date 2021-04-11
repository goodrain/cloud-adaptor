// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/jinzhu/gorm"
	"goodrain.com/cloud-adaptor/internal/biz"
	"goodrain.com/cloud-adaptor/internal/data"
	"goodrain.com/cloud-adaptor/internal/handler"
	"goodrain.com/cloud-adaptor/internal/nsqc/producer"
)

// initGin init gin engine..
func initGin(*gorm.DB, producer.TaskProducer) (*gin.Engine, error) {
	panic(wire.Build(handler.ProviderSet, biz.ProviderSet, data.ProviderSet, newGinEngine))
}
