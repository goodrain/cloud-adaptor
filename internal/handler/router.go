// RAINBOND, Application Management Platform
// Copyright (C) 2020-2021 Goodrain Co., Ltd.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. For any non-GPL usage of Rainbond,
// one or multiple Commercial Licenses authorized by Goodrain Co., Ltd.
// must be obtained first.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"goodrain.com/cloud-adaptor/internal/middleware"
	"goodrain.com/cloud-adaptor/pkg/util/constants"

	// go-swag
	_ "goodrain.com/cloud-adaptor/docs"
)

// Router -
type Router struct {
	middleware *middleware.Middleware
	cluster    *ClusterHandler
	system     *SystemHandler
	appStore   *AppStoreHandler
	helm       *HelmHandler
}

// NewRouter creates a new router.
func NewRouter(
	middleware *middleware.Middleware,
	cluster *ClusterHandler,
	appStore *AppStoreHandler,
	system *SystemHandler,
) *Router {
	return &Router{
		middleware: middleware,
		cluster:    cluster,
		appStore:   appStore,
		system:     system,
	}
}

// SetCORS Enables cross-site script calls.
func SetCORS(ctx *gin.Context) {
	origin := ctx.GetHeader("Origin")
	ctx.Writer.Header().Add("Access-Control-Allow-Origin", origin)
	ctx.Writer.Header().Add("Access-Control-Allow-Methods", "POST,GET,OPTIONS,DELETE,PUT")
	ctx.Writer.Header().Add("Access-Control-Allow-Credentials", "true")
	ctx.Writer.Header().Add("Access-Control-Allow-Headers", "x-requested-with,content-type,Authorization,X-Token")
}

// CORSMidle -
var CORSMidle = func(f gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		SetCORS(ctx)
		f(ctx)
	}
}

// NewRouter creates a new Router
func (r *Router) NewRouter() *gin.Engine {
	gin.SetMode(gin.DebugMode)
	e := gin.Default()
	e.OPTIONS("/*path", CORSMidle(func(ctx *gin.Context) {}))

	g := e.Group(constants.Service)
	// openapi
	apiv1 := g.Group("/api/v1")
	apiv1.GET("/backup", r.system.Backup)
	apiv1.POST("/recover", r.system.Recover)
	apiv1.GET("/init_node_cmd", r.cluster.GetInitNodeCmd)
	apiv1.POST("/check_ssh", r.cluster.CheckSSH)

	apiv1.POST("/helm/chart", CORSMidle(r.helm.GetHelmCommand))
	entv1 := apiv1.Group("/enterprises/:eid")
	// cluster
	entv1.GET("/kclusters", r.cluster.ListKubernetesClusters)
	entv1.POST("/kclusters", r.cluster.AddKubernetesCluster)
	entv1.GET("/kclusters/:clusterID/regionconfig", r.cluster.GetRegionConfig)
	entv1.DELETE("/kclusters/:clusterID", r.cluster.DeleteKubernetesCluster)
	entv1.POST("/kclusters/:clusterID/reinstall", r.cluster.ReInstallKubernetesCluster)
	entv1.GET("/kclusters/:clusterID/createlog", r.cluster.GetLogContent)
	entv1.GET("/kclusters/:clusterID/kubeconfig", r.cluster.GetKubeConfig)
	entv1.GET("/kclusters/:clusterID/rainbondcluster", r.cluster.GetRainbondClusterConfig)
	entv1.PUT("/kclusters/:clusterID/rainbondcluster", r.cluster.SetRainbondClusterConfig)
	entv1.POST("/kclusters/:clusterID/uninstall", r.cluster.UninstallRegion)
	entv1.POST("/kclusters/prune-update-rkeconfig", r.cluster.pruneUpdateRKEConfig)

	clusterv1 := entv1.Group("/kclusters/:clusterID")
	{
		clusterv1.GET("/rainbond-components", r.cluster.listRainbondComponents)
		clusterv1.GET("/rainbond-components/:podName/events", r.cluster.listPodEvents)
	}

	entv1.POST("/accesskey", r.cluster.AddAccessKey)
	entv1.GET("/accesskey", r.cluster.GetAccessKey)
	entv1.GET("/last-ck-task", r.cluster.GetLastAddKubernetesClusterTask)
	entv1.GET("/ck-task/:taskID", r.cluster.GetAddKubernetesClusterTask)
	entv1.GET("/tasks/:taskID/events", r.cluster.GetTaskEventList)
	entv1.GET("/init-task/:clusterID", r.cluster.GetInitRainbondTask)
	entv1.GET("/init-tasks", r.cluster.GetRunningInitRainbondTask)
	entv1.POST("/init-cluster", r.cluster.CreateInitRainbondTask)
	entv1.PUT("/init-tasks/:taskID/status", r.cluster.UpdateInitRainbondTaskStatus)

	entv1.POST("/update-cluster", r.cluster.UpdateKubernetesCluster)
	entv1.GET("/update-cluster/:clusterID", r.cluster.GetUpdateKubernetesTask)

	// app store
	appstoresv1 := entv1.Group("/appstores")
	appstoresv1.POST("", r.appStore.Create)
	appstoresv1.GET("", r.appStore.List)
	appstorev1 := appstoresv1.Group(":name", r.middleware.AppStore)
	{
		appstorev1.GET("", r.appStore.Get)
		appstorev1.PUT("", r.appStore.Update)
		appstorev1.DELETE("", r.appStore.Delete)
		appstorev1.POST("/resync", r.appStore.Resync)
		// TODO: change app to templates
		appstorev1.GET("/apps", r.appStore.ListTemplates)
		appstorev1.GET("/apps/:templateName", r.appStore.GetAppTemplate)
		appstorev1.GET("/templates/:templateName/versions/:version", r.appStore.GetAppTemplateVersion)
	}

	e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return e
}
