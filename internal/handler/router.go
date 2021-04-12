package handler

import (
	"github.com/gin-gonic/gin"
	"goodrain.com/cloud-adaptor/pkg/util/constants"
)

// Router -
type Router struct {
	ClusterHandler *ClusterHandler
	//SystemHandler  *oahan.SystemHandler
	appStore *AppStoreHandler
}

// NewRouter creates a new router.
func NewRouter(cluster *ClusterHandler,
	appStore *AppStoreHandler) *Router {
	return &Router{
		ClusterHandler: cluster,
		appStore:       appStore,
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
	//apiv1.GET("/backup", r.SystemHandler.Backup)
	//apiv1.POST("/recover", r.SystemHandler.Recover)
	apiv1.GET("/init_node_cmd", r.ClusterHandler.GetInitNodeCmd)
	entv1 := apiv1.Group("/enterprises/:eid")
	// cluster
	entv1.GET("/kclusters", r.ClusterHandler.ListKubernetesClusters)
	entv1.GET("/kclusters/:clusterID/regionconfig", r.ClusterHandler.GetRegionConfig)
	entv1.POST("/kclusters", r.ClusterHandler.AddKubernetesCluster)
	entv1.DELETE("/kclusters/:clusterID", r.ClusterHandler.DeleteKubernetesCluster)
	entv1.POST("/kclusters/:clusterID/reinstall", r.ClusterHandler.ReInstallKubernetesCluster)
	entv1.GET("/kclusters/:clusterID/createlog", r.ClusterHandler.GetLogContent)
	entv1.GET("/kclusters/:clusterID/kubeconfig", r.ClusterHandler.GetKubeConfig)
	entv1.GET("/kclusters/:clusterID/rainbondcluster", r.ClusterHandler.GetRainbondClusterConfig)
	entv1.PUT("/kclusters/:clusterID/rainbondcluster", r.ClusterHandler.SetRainbondClusterConfig)
	entv1.POST("/kclusters/:clusterID/uninstall", r.ClusterHandler.UninstallRegion)

	entv1.POST("/accesskey", r.ClusterHandler.AddAccessKey)
	entv1.GET("/accesskey", r.ClusterHandler.GetAccessKey)
	entv1.GET("/last-ck-task", r.ClusterHandler.GetLastAddKubernetesClusterTask)
	entv1.GET("/ck-task/:taskID", r.ClusterHandler.GetAddKubernetesClusterTask)
	entv1.GET("/tasks/:taskID/events", r.ClusterHandler.GetTaskEventList)
	entv1.GET("/init-task/:clusterID", r.ClusterHandler.GetInitRainbondTask)
	entv1.GET("/init-tasks", r.ClusterHandler.GetRunningInitRainbondTask)
	entv1.POST("/init-cluster", r.ClusterHandler.CreateInitRainbondTask)
	entv1.PUT("/init-tasks/:taskID/status", r.ClusterHandler.UpdateInitRainbondTaskStatus)

	entv1.POST("/update-cluster", r.ClusterHandler.UpdateKubernetesCluster)
	entv1.GET("/update-cluster/:clusterID", r.ClusterHandler.GetUpdateKubernetesTask)

	// app store
	appstorev1 := entv1.Group("/appstores")
	appstorev1.POST("/", r.appStore.Create)

	return e
}
