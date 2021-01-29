package router

import (
	"github.com/gin-gonic/gin"

	"goodrain.com/cloud-adaptor/api/config"
	oahan "goodrain.com/cloud-adaptor/api/openapi/handler"
	"goodrain.com/cloud-adaptor/util/constants"
)

// Router -
type Router struct {
	Config         *config.Config        `inject:""`
	ClusterHandler *oahan.ClusterHandler `inject:""`
}

// New creates a new router.
func New() *Router {
	return &Router{}
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
	return gin.HandlerFunc(func(ctx *gin.Context) {
		SetCORS(ctx)
		f(ctx)
	})
}

// NewRouter creates a new Router
func (r *Router) NewRouter() *gin.Engine {
	e := gin.Default()
	e.OPTIONS("/*path", CORSMidle(func(ctx *gin.Context) {}))

	g := e.Group(constants.Service)
	// openapi
	apiv1 := g.Group("/api/v1")
	apiv1.GET("/init_node_cmd", r.ClusterHandler.GetInitNodeCmd)
	entv1 := apiv1.Group("/enterprises")
	// cluster
	entv1.GET("/:eid/kclusters", r.ClusterHandler.ListKubernetesClusters)
	entv1.GET("/:eid/kclusters/:clusterID/regionconfig", r.ClusterHandler.GetRegionConfig)
	entv1.POST("/:eid/kclusters", r.ClusterHandler.AddKubernetesCluster)
	entv1.DELETE("/:eid/kclusters/:clusterID", r.ClusterHandler.DeleteKubernetesCluster)
	entv1.POST("/:eid/kclusters/:clusterID/reinstall", r.ClusterHandler.ReInstallKubernetesCluster)
	entv1.GET("/:eid/kclusters/:clusterID/createlog", r.ClusterHandler.GetLogContent)
	entv1.GET("/:eid/kclusters/:clusterID/kubeconfig", r.ClusterHandler.GetKubeConfig)

	entv1.POST("/:eid/accesskey", r.ClusterHandler.AddAccessKey)
	entv1.GET("/:eid/accesskey", r.ClusterHandler.GetAccessKey)
	entv1.GET("/:eid/last-ck-task", r.ClusterHandler.GetLastAddKubernetesClusterTask)
	entv1.GET("/:eid/ck-task/:taskID", r.ClusterHandler.GetAddKubernetesClusterTask)
	entv1.GET("/:eid/tasks/:taskID/events", r.ClusterHandler.GetTaskEventList)
	entv1.GET("/:eid/init-task/:clusterID", r.ClusterHandler.GetInitRainbondTask)
	entv1.GET("/:eid/init-tasks", r.ClusterHandler.GetRunningInitRainbondTask)
	entv1.POST("/:eid/init-cluster", r.ClusterHandler.CreateInitRainbondTask)
	entv1.PUT("/:eid/init-tasks/:taskID/status", r.ClusterHandler.UpdateInitRainbondTaskStatus)

	entv1.POST("/:eid/update-cluster", r.ClusterHandler.UpdateKubernetesCluster)
	entv1.GET("/:eid/update-cluster/:clusterID", r.ClusterHandler.GetUpdateKubernetesTask)
	return e
}
