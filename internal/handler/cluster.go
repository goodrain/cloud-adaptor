// RAINBOND, Application Management Platform
// Copyright (C) 2020-2020 Goodrain Co., Ltd.

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
	"goodrain.com/cloud-adaptor/pkg/util/ssh"
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "goodrain.com/cloud-adaptor/api/cloud-adaptor/v1"
	"goodrain.com/cloud-adaptor/internal/adaptor/v1alpha1"
	"goodrain.com/cloud-adaptor/internal/usecase"
	"goodrain.com/cloud-adaptor/pkg/bcode"
	"goodrain.com/cloud-adaptor/pkg/util/ginutil"
	"goodrain.com/cloud-adaptor/pkg/util/md5util"
)

// ClusterHandler -
type ClusterHandler struct {
	cluster *usecase.ClusterUsecase
}

// NewClusterHandler new enterprise handler
func NewClusterHandler(clusterUsecase *usecase.ClusterUsecase) *ClusterHandler {
	return &ClusterHandler{
		cluster: clusterUsecase,
	}
}

// ListKubernetesClusters returns the information of .
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/kclusters cloud kcluster
//
// # ListKubernetesCluster
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:KubernetesClustersResponse
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) ListKubernetesClusters(ctx *gin.Context) {
	var req v1.ListKubernetesCluster
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logrus.Errorf("bind query param failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	eid := ctx.Param("eid")
	clusters, err := e.cluster.ListKubernetesCluster(eid, req)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, v1.KubernetesClustersResponse{Clusters: clusters}, nil)
}

// AddKubernetesCluster returns the information of .
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/kclusters cloud kcluster
//
// # CreateKubernetesReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:CreateKubernetesRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) AddKubernetesCluster(ctx *gin.Context) {
	var req v1.CreateKubernetesReq
	if err := ginutil.ShouldBindJSON(ctx, &req); err != nil {
		ginutil.Error(ctx, err)
		return
	}
	if req.Provider == "rke" {
		if req.EncodedRKEConfig == "" {
			ginutil.JSON(ctx, nil, bcode.ErrIncorrectRKEConfig)
			return
		}
	}
	if req.Provider == "custom" {
		if req.KubeConfig == "" {
			ginutil.JSON(ctx, nil, bcode.ErrKubeConfigCannotEmpty)
			return
		}
	}
	eid := ctx.Param("eid")
	task, err := e.cluster.CreateKubernetesCluster(eid, req)
	if err != nil {
		ginutil.JSON(ctx, task, err)
		return
	}
	ginutil.JSON(ctx, task, nil)
}

// UpdateKubernetesCluster updates kubernetes cluster.
//
// @Summary updates kubernetes cluster.
// @Tags cluster
// @ID updateKubernetesCluster
// @Accept  json
// @Produce  json
// @Param eid path string true "the enterprise id"
// @Param updateKubernetesReq body v1.UpdateKubernetesReq true "."
// @Success 200 {object} v1.UpdateKubernetesTask
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/enterprises/:eid/update-cluster [post]
func (e *ClusterHandler) UpdateKubernetesCluster(ctx *gin.Context) {
	var req v1.UpdateKubernetesReq
	if err := ginutil.ShouldBindJSON(ctx, &req); err != nil {
		ginutil.Error(ctx, err)
		return
	}
	if req.Provider == "rke" {
		if req.EncodedRKEConfig == "" {
			ginutil.Error(ctx, errors.WithMessage(bcode.ErrIncorrectRKEConfig, "rke config is required"))
			return
		}
	}
	eid := ctx.Param("eid")
	task, err := e.cluster.UpdateKubernetesCluster(eid, req)
	if err != nil {
		ginutil.JSONv2(ctx, task, err)
		return
	}
	ginutil.JSONv2(ctx, task)
}

// GetUpdateKubernetesTask returns the information of the cluster.
//
// @Summary  returns the information of the cluster.
// @Tags cluster
// @ID getUpdateKubernetesTask
// @Accept  json
// @Produce  json
// @Param eid path string true "the enterprise id"
// @Param clusterID path string true "the cluster id"
// @Success 200 {object} v1.UpdateKubernetesTask
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/enterprises/:eid/update-cluster/:clusterID [get]
func (e *ClusterHandler) GetUpdateKubernetesTask(ctx *gin.Context) {
	eid := ctx.Param("eid")
	clusterID := ctx.Param("clusterID")
	providerName := ctx.Query("provider_name")
	re, err := e.cluster.GetUpdateKubernetesTask(eid, clusterID, providerName)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}

	ginutil.JSON(ctx, re, nil)
}

// DeleteKubernetesCluster returns the information of .
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/kclusters/{clusterID} cloud kcluster
//
// # DeleteKubernetesClusterReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:Reponse
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) DeleteKubernetesCluster(ctx *gin.Context) {
	var req v1.DeleteKubernetesClusterReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logrus.Errorf("bind query param failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	eid := ctx.Param("eid")
	clusterID := ctx.Param("clusterID")
	err := e.cluster.DeleteKubernetesCluster(eid, clusterID, req.ProviderName)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, nil, nil)
}

// GetLastAddKubernetesClusterTask returns the information of .
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/last-ck-task cloud kcluster
//
// # GetLastCreateKubernetesClusterTaskReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:GetCreateKubernetesClusterTaskRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) GetLastAddKubernetesClusterTask(ctx *gin.Context) {
	var req v1.GetLastCreateKubernetesClusterTaskReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logrus.Errorf("bind query param failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	eid := ctx.Param("eid")
	task, err := e.cluster.GetLastCreateKubernetesTask(eid, req.ProviderName)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, task, nil)
}

// GetAddKubernetesClusterTask returns the information of .
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/ck-task/{taskID} cloud kcluster
//
// # GetLastCreateKubernetesClusterTaskReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:GetCreateKubernetesClusterTaskRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) GetAddKubernetesClusterTask(ctx *gin.Context) {
	eid := ctx.Param("eid")
	taskID := ctx.Param("taskID")
	task, err := e.cluster.GetCreateKubernetesTask(eid, taskID)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, task, nil)
}

// GetTaskEventList returns the information of .
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/ck-task/{taskID}/events cloud kcluster
//
// # GetTaskEventListReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:GetCreateKubernetesClusterTaskRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) GetTaskEventList(ctx *gin.Context) {
	eid := ctx.Param("eid")
	taskID := ctx.Param("taskID")
	events, err := e.cluster.ListTaskEvent(eid, taskID)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, v1.TaskEventListRes{Events: events}, nil)
}

// AddAccessKey add access keys
func (e *ClusterHandler) AddAccessKey(ctx *gin.Context) {
	var req v1.AddAccessKey
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("bind add accesskey param failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	eid := ctx.Param("eid")
	clusters, err := e.cluster.AddAccessKey(eid, req)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, clusters, nil)
}

// GetAccessKey add access keys
func (e *ClusterHandler) GetAccessKey(ctx *gin.Context) {
	var req v1.GetAccessKeyReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logrus.Errorf("bind add accesskey param failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	eid := ctx.Param("eid")
	access, err := e.cluster.GetByProviderAndEnterprise(req.ProviderName, eid)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	access.SecretKey = md5util.Md5Crypt(access.SecretKey, access.EnterpriseID)
	ginutil.JSON(ctx, access, nil)
}

// GetInitRainbondTask returns the information of .
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/init-task/{clusterID} cloud init
//
// # GetInitRainbondTaskReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:InitRainbondTaskRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) GetInitRainbondTask(ctx *gin.Context) {
	eid := ctx.Param("eid")
	clusterID := ctx.Param("clusterID")
	var req v1.GetInitRainbondTaskReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logrus.Errorf("bind get init rainbond task query failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	task, err := e.cluster.GetInitRainbondTaskByClusterID(eid, clusterID, req.ProviderName)
	ginutil.JSON(ctx, task, err)
}

// CreateInitRainbondTask returns the information of .
//
// swagger:route POST /enterprise-server/api/v1/enterprises/{eid}/init-cluster cloud init
//
// # InitRainbondRegionReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:InitRainbondTaskRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) CreateInitRainbondTask(ctx *gin.Context) {
	eid := ctx.Param("eid")
	var req v1.InitRainbondRegionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("bind init rainbond body failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	task, err := e.cluster.InitRainbondRegion(ctx.Request.Context(), eid, req)
	if err != nil {
		ginutil.JSON(ctx, task, err)
		return
	}
	ginutil.JSON(ctx, task, nil)
}

// GetRunningInitRainbondTask returns the information of .
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/init-task/{clusterID} cloud init
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:InitRainbondTaskListRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) GetRunningInitRainbondTask(ctx *gin.Context) {
	eid := ctx.Param("eid")
	tasks, err := e.cluster.GetTaskRunningLists(eid)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, v1.InitRainbondTaskListRes{Tasks: tasks}, nil)
}

// GetRegionConfig get region config file
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/kclusters/{clusterID}/regionconfig cloud kcluster
//
// # GetRegionConfigReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:GetRegionConfigRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) GetRegionConfig(ctx *gin.Context) {
	var req v1.GetRegionConfigReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logrus.Errorf("bind get rainbond region config failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	eid := ctx.Param("eid")
	clusterID := ctx.Param("clusterID")
	configs, err := e.cluster.GetRegionConfig(eid, clusterID, req.ProviderName)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	out, _ := yaml.Marshal(configs)
	ginutil.JSON(ctx, v1.GetRegionConfigRes{Configs: configs, ConfigYaml: string(out)}, nil)
}

// UpdateInitRainbondTaskStatus get region config file
//
// swagger:route PUT /enterprise-server/api/v1/enterprises/{eid}/init-tasks/{taskID}/status cloud init
//
// # UpdateInitRainbondTaskStatusReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:InitRainbondTaskRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) UpdateInitRainbondTaskStatus(ctx *gin.Context) {
	var req v1.UpdateInitRainbondTaskStatusReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("bind update init status failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	eid := ctx.Param("eid")
	taskID := ctx.Param("taskID")
	task, err := e.cluster.UpdateInitRainbondTaskStatus(eid, taskID, req.Status)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, task, nil)
}

// GetInitNodeCmd get node init cmd shell
//
// swagger:route GET /enterprise-server/api/v1/init_node_cmd cloud init
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:InitNodeCmdRes
func (e *ClusterHandler) GetInitNodeCmd(c *gin.Context) {
	res, err := e.cluster.GetInitNodeCmd(c.Request.Context())
	ginutil.JSONv2(c, res, err)
}

// check ssh connect
//
// swagger:route GET /enterprise-server/api/v1/check_ssh
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:bool
func (e *ClusterHandler) CheckSSH(ctx *gin.Context) {
	var req v1.CheckSSHReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	r, err := ssh.CheckSSHConnect(req.Host, req.Port)
	if err != nil {
		ginutil.JSON(ctx, r, err)
		return
	}
	var res = v1.CheckSSHRes{
		Status: r,
	}
	ginutil.JSON(ctx, res)
}

// GetLogContent get rke create kubernetes log
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/kclusters/{clusterID}/create_log cloud init
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:GetLogContentRes
func (e *ClusterHandler) GetLogContent(ctx *gin.Context) {
	cluster, err := e.cluster.GetCluster("rke", ctx.Param("eid"), ctx.Param("clusterID"))
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	var content []byte
	if cluster.CreateLogPath != "" {
		content, _ = ioutil.ReadFile(cluster.CreateLogPath)
	}
	ginutil.JSON(ctx, v1.GetLogContentRes{Content: string(content)}, nil)
}

// ReInstallKubernetesCluster retry install rke cluster .
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/kclusters/{clusterID}/reinstall cloud kcluster
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:CreateKubernetesRes
// 400: body:Reponse
// 500: body:Reponse
func (e *ClusterHandler) ReInstallKubernetesCluster(ctx *gin.Context) {
	task, err := e.cluster.InstallCluster(ctx.Param("eid"), ctx.Param("clusterID"))
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, task, nil)
}

// GetKubeConfig get kubernetes cluster config
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/kclusters/{clusterID}/kubeconfig cloud init
//
// # GetRegionConfigReq
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:GetKubeConfigRes
func (e *ClusterHandler) GetKubeConfig(ctx *gin.Context) {
	var req v1.GetRegionConfigReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logrus.Errorf("bind get rainbond region config failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	kubeconfig, err := e.cluster.GetKubeConfig(ctx.Param("eid"), ctx.Param("clusterID"), req.ProviderName)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			ginutil.JSON(ctx, nil, bcode.NotFound)
		}
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, v1.GetKubeConfigRes{Config: kubeconfig}, nil)
}

// GetRainbondClusterConfig -
func (e *ClusterHandler) GetRainbondClusterConfig(ctx *gin.Context) {
	eid := ctx.Param("eid")
	clusterID := ctx.Param("clusterID")
	_, config := e.cluster.GetRainbondClusterConfig(eid, clusterID)
	if config == "" {
		config = `
# apiVersion: rainbond.io/v1alpha1
# kind: RainbondCluster
# metadata:
#  name: rainbondcluster
#  namespace: rbd-system
# spec:
#  ## set source build cache mode, default is hostpath, options: pv, hostpath
#  cacheMode: hostpath
#  configCompleted: true
#  ## Whether to deploy high availability. default is true if the number of nodes is greater than 3.
#  enableHA: false
#  ## etcd config, secret must have ca-file、cert-file and key-file keys.
#  etcdConfig:
#	 endpoints:
#	 - 192.168.10.6:2379
#	 - 192.168.10.8:2379
#	 - 192.168.10.4:2379
#	 secretName: rbd-etcd-secret
#  ## Specifies the outer network IP address of the gateway. As the access address.
#  gatewayIngressIPs:
#    - 39.101.149.237
#  ## Specifies image hub info, deployment default hub when not set.
#  imageHub:
#	 domain: goodrain.me
#	 password: 526856c5
#	 username: admin
#  installVersion: v5.3.0-release
#  ## Specifies the node that performs the component CI task.
#  nodesForChaos:
#   - externalIP: 121.89.192.53
#	  internalIP: 192.168.10.3
#	  name: 39.101.149.237
#  ## Specify the gateway node.
#  nodesForGateway:
#   - externalIP: 121.89.192.53
#	  internalIP: 192.168.10.3
#	  name: 39.101.149.237
#  ## Specifies the rainbond component image hub address
#  rainbondImageRepository: registry.cn-hangzhou.aliyuncs.com/goodrain
#  ## Specifies shared storage provider.
#  rainbondVolumeSpecRWX:
#	 imageRepository: ""
#	 storageClassName: glusterfs-simple
#  ## Specifies the db connection info of region.
#  regionDatabase:
#	 host: 127.0.0.1
#	 name: region
#	 password: rainbond123456!
#	 port: 3306
#	 username: root
#  ## Specifies the default component domain name suffix. Not specified will be assigned by default
#  suffixHTTPHost: xxxx.grapps.cn`
	}
	re := v1.SetRainbondClusterConfigReq{
		Config: config,
	}
	ginutil.JSON(ctx, re, nil)
}

// SetRainbondClusterConfig -
func (e *ClusterHandler) SetRainbondClusterConfig(ctx *gin.Context) {
	eid := ctx.Param("eid")
	clusterID := ctx.Param("clusterID")
	var req v1.SetRainbondClusterConfigReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("bind update init status failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	err := e.cluster.SetRainbondClusterConfig(eid, clusterID, req.Config)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, nil, nil)
}

// UninstallRegion -
func (e *ClusterHandler) UninstallRegion(ctx *gin.Context) {
	eid := ctx.Param("eid")
	clusterID := ctx.Param("clusterID")
	var req v1.UninstallRegionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("bind update init status failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	err := e.cluster.UninstallRainbondRegion(eid, clusterID, req.ProviderName)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, nil, nil)
}

// @Summary update rke config purely
// @Tags cluster
// @ID pruneUpdateRKEConfig
// @Accept  json
// @Produce  json
// @Param eid path string true "the enterprise id"
// @Param pruneUpdateRKEConfigReq body v1.PruneUpdateRKEConfigReq true "."
// @Success 200 {object} v1.PruneUpdateRKEConfigResp
// @Failure 500 {object} ginutil.Result
// @Router /api/v1/enterprises/:eid/kclusters/prune-update-rkeconfig [POST]
func (e *ClusterHandler) pruneUpdateRKEConfig(c *gin.Context) {
	var req v1.PruneUpdateRKEConfigReq
	if err := ginutil.ShouldBindJSON(c, &req); err != nil {
		ginutil.Error(c, err)
		return
	}

	// clean invalid nodes
	var nodes v1alpha1.NodeList
	for _, node := range req.Nodes {
		if node.IP == "" || node.InternalAddress == "" {
			continue
		}
		nodes = append(nodes, node)
	}
	req.Nodes = nodes

	rkeConfig, err := e.cluster.PruneUpdateRKEConfig(&req)
	ginutil.JSONv2(c, rkeConfig, err)
}

// ListRainbondComponents returns a list of rainbond components.
// @Summary returns a list of rainbond components.
// @Tags cluster
// @ID listRainbondComponents
// @Accept  json
// @Produce  json
// @Param eid path string true "the enterprise id"
// @Param clusterID path string true "the identify of cluster"
// @Param providerName query string true "the provider of the cluster"
// @Success 200 {array} v1.RainbondComponent
// @Router /api/v1/enterprises/{eid}/kclusters/{clusterID}/rainbond-components [get]
func (e *ClusterHandler) listRainbondComponents(c *gin.Context) {
	eid := c.Param("eid")
	clusterID := c.Param("clusterID")
	providerName := c.Query("providerName")
	components, err := e.cluster.ListRainbondComponents(c.Request.Context(), eid, clusterID, providerName)
	ginutil.JSONv2(c, components, err)
}

// listPodEvents returns a list of rainbond component pod events.
// @Summary returns a list of rainbond component pod events.
// @Tags cluster
// @ID listPodEvents
// @Accept  json
// @Produce  json
// @Param eid path string true "the enterprise id"
// @Param clusterID path string true "the identify of cluster"
// @Param podName path string true "the name of pod"
// @Param providerName query string true "the provider of the cluster"
// @Success 200 {array} v1.RainbondComponentEvent
// @Router /api/v1/enterprises/{eid}/kclusters/{clusterID}/rainbond-components/{podName}/events [get]
func (e *ClusterHandler) listPodEvents(c *gin.Context) {
	eid := c.Param("eid")
	clusterID := c.Param("clusterID")
	providerName := c.Query("providerName")
	components, err := e.cluster.ListPodEvents(c.Request.Context(), eid, clusterID, providerName, c.Param("podName"))
	ginutil.JSONv2(c, components, err)
}
