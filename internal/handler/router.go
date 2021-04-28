package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "goodrain.com/cloud-adaptor/docs"
	"goodrain.com/cloud-adaptor/internal/middleware"
	"goodrain.com/cloud-adaptor/pkg/util/constants"
)

// Router -
type Router struct {
	middleware *middleware.Middleware
	cluster    *ClusterHandler
	system     *SystemHandler
	appStore   *AppStoreHandler
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

//SetCORS Enables cross-site script calls.
func SetCORS(ctx *gin.Context) {
	origin := ctx.GetHeader("Origin")
	ctx.Writer.Header().Add("Access-Control-Allow-Origin", origin)
	ctx.Writer.Header().Add("Access-Control-Allow-Methods", "POST,GET,OPTIONS,DELETE,PUT")
	ctx.Writer.Header().Add("Access-Control-Allow-Credentials", "true")
	ctx.Writer.Header().Add("Access-Control-Allow-Headers", "x-requested-with,content-type,Authorization,X-Token")
}

//CORSMidle -
var CORSMidle = func(f gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		SetCORS(ctx)
		f(ctx)
	}
}

// NewRouter creates a new Router
func (r *Router) NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	e := gin.Default()
	e.OPTIONS("/*path", CORSMidle(func(ctx *gin.Context) {}))

	g := e.Group(constants.Service)
	// openapi
	apiv1 := g.Group("/api/v1")
	apiv1.GET("/backup", r.system.Backup)
	apiv1.POST("/recover", r.system.Recover)
	apiv1.GET("/init_node_cmd", r.cluster.GetInitNodeCmd)
	entv1 := apiv1.Group("/enterprises/:eid")
	// cluster
	entv1.GET("/kclusters", r.cluster.ListKubernetesClusters)
	entv1.GET("/kclusters/:clusterID/regionconfig", r.cluster.GetRegionConfig)
	entv1.POST("/kclusters", r.cluster.AddKubernetesCluster)
	entv1.DELETE("/kclusters/:clusterID", r.cluster.DeleteKubernetesCluster)
	entv1.POST("/kclusters/:clusterID/reinstall", r.cluster.ReInstallKubernetesCluster)
	entv1.GET("/kclusters/:clusterID/createlog", r.cluster.GetLogContent)
	entv1.GET("/kclusters/:clusterID/kubeconfig", r.cluster.GetKubeConfig)
	entv1.GET("/kclusters/:clusterID/rainbondcluster", r.cluster.GetRainbondClusterConfig)
	entv1.PUT("/kclusters/:clusterID/rainbondcluster", r.cluster.SetRainbondClusterConfig)
	entv1.POST("/kclusters/:clusterID/uninstall", r.cluster.UninstallRegion)

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
		appstorev1.GET("/apps", r.appStore.ListTemplates)
		appstorev1.GET("/apps/:templateName", r.appStore.GetAppTemplate)
	}

	e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return e
}
