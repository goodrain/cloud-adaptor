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
	"fmt"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"goodrain.com/cloud-adaptor/api/cluster"
	v1 "goodrain.com/cloud-adaptor/api/openapi/types/v1"
	"goodrain.com/cloud-adaptor/library/bcode"
	"goodrain.com/cloud-adaptor/util/ginutil"
	"goodrain.com/cloud-adaptor/util/ssh"
	yaml "gopkg.in/yaml.v2"
)

// ClusterHandler -
type ClusterHandler struct {
	ClusterUsecase cluster.Usecase `inject:""`
}

// NewClusterHandler new enterprise handler
func NewClusterHandler() *ClusterHandler {
	return &ClusterHandler{}
}

// ListKubernetesClusters returns the information of .
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/kclusters cloud kcluster
//
// ListKubernetesCluster
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
	clusters, err := e.ClusterUsecase.ListKubernetesCluster(eid, req)
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
// CreateKubernetesReq
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
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("bind body param failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	if req.Provider == "rke" {
		if err := req.Nodes.Validate(); err != nil {
			ginutil.JSON(ctx, nil, err)
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
	task, err := e.ClusterUsecase.CreateKubernetesCluster(eid, req)
	if err != nil {
		ginutil.JSON(ctx, task, err)
		return
	}
	ginutil.JSON(ctx, task, nil)
}

// DeleteKubernetesCluster returns the information of .
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/kclusters/{clusterID} cloud kcluster
//
// DeleteKubernetesClusterReq
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
	err := e.ClusterUsecase.DeleteKubernetesCluster(eid, clusterID, req.ProviderName)
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
// GetLastCreateKubernetesClusterTaskReq
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
	task, err := e.ClusterUsecase.GetLastCreateKubernetesTask(eid, req.ProviderName)
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
// GetLastCreateKubernetesClusterTaskReq
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
	task, err := e.ClusterUsecase.GetCreateKubernetesTask(eid, taskID)
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
// GetTaskEventListReq
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
	events, err := e.ClusterUsecase.ListTaskEvent(eid, taskID)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, v1.TaskEventListRes{Events: events}, nil)
}

//AddAccessKey add access keys
func (e *ClusterHandler) AddAccessKey(ctx *gin.Context) {
	var req v1.AddAccessKey
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("bind add accesskey param failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	eid := ctx.Param("eid")
	clusters, err := e.ClusterUsecase.AddAccessKey(eid, req)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, clusters, nil)
}

//GetAccessKey add access keys
func (e *ClusterHandler) GetAccessKey(ctx *gin.Context) {
	var req v1.GetAccessKeyReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logrus.Errorf("bind add accesskey param failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.BadRequest)
		return
	}
	eid := ctx.Param("eid")
	access, err := e.ClusterUsecase.GetByProviderAndEnterprise(req.ProviderName, eid)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, access, nil)
}

// GetInitRainbondTask returns the information of .
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/init-task/{clusterID} cloud init
//
// GetInitRainbondTaskReq
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
	task, err := e.ClusterUsecase.GetInitRainbondTaskByClusterID(eid, clusterID, req.ProviderName)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, task, nil)
}

// CreateInitRainbondTask returns the information of .
//
// swagger:route POST /enterprise-server/api/v1/enterprises/{eid}/init-cluster cloud init
//
// InitRainbondRegionReq
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
	task, err := e.ClusterUsecase.InitRainbondRegion(eid, req)
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
	tasks, err := e.ClusterUsecase.GetTaskRunningLists(eid)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, v1.InitRainbondTaskListRes{Tasks: tasks}, nil)
}

//GetRegionConfig get region config file
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/kclusters/{clusterID}/regionconfig cloud kcluster
//
// GetRegionConfigReq
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
	configs, err := e.ClusterUsecase.GetRegionConfig(eid, clusterID, req.ProviderName)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	out, _ := yaml.Marshal(configs)
	ginutil.JSON(ctx, v1.GetRegionConfigRes{Configs: configs, ConfigYaml: string(out)}, nil)
}

//UpdateInitRainbondTaskStatus get region config file
//
// swagger:route PUT /enterprise-server/api/v1/enterprises/{eid}/init-tasks/{taskID}/status cloud init
//
// UpdateInitRainbondTaskStatusReq
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
	task, err := e.ClusterUsecase.UpdateInitRainbondTaskStatus(eid, taskID, req.Status)
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, task, nil)
}

//GetInitNodeCmd get node init cmd shell
//
// swagger:route GET /enterprise-server/api/v1/init_node_cmd cloud init
//
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:InitNodeCmdRes
func (e *ClusterHandler) GetInitNodeCmd(ctx *gin.Context) {
	pub, err := ssh.GetOrMakeSSHRSA()
	if err != nil {
		logrus.Errorf("get or create ssh rsa failure %s", err.Error())
		ginutil.JSON(ctx, nil, bcode.ServerErr)
		return
	}
	ginutil.JSON(ctx, v1.InitNodeCmdRes{
		Cmd: fmt.Sprintf(`export SSH_RSA="%s"&&curl http://sh.rainbond.com/init_node | bash`, pub),
	}, nil)
}

//GetLogContent get rke create kubernetes log
//
// swagger:route GET /enterprise-server/api/v1/enterprises/{eid}/kclusters/{clusterID}/create_log cloud init
//
//
// Produces:
// - application/json
// Schemes: http
// Consumes:
// - application/json
//
// Responses:
// 200: body:string
func (e *ClusterHandler) GetLogContent(ctx *gin.Context) {
	cluster, err := e.ClusterUsecase.GetCluster("rke", ctx.Param("eid"), ctx.Param("clusterID"))
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
	task, err := e.ClusterUsecase.InstallCluster(ctx.Param("eid"), ctx.Param("clusterID"))
	if err != nil {
		ginutil.JSON(ctx, nil, err)
		return
	}
	ginutil.JSON(ctx, task, nil)
}
