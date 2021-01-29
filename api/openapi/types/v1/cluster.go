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

package v1

import (
	"goodrain.com/cloud-adaptor/adaptor/v1alpha1"
	"goodrain.com/cloud-adaptor/api/models"
)

//ListKubernetesCluster list kubernetes cluster request body
//swagger:model ListKubernetesCluster
type ListKubernetesCluster struct {
	ProviderName string `form:"provider_name" binding:"required"`
}

//AddAccessKey -
//swagger:model AddAccessKey
type AddAccessKey struct {
	ProviderName string `json:"provider_name,omitempty" binding:"required"`
	AccessKey    string `json:"access_key,omitempty" binding:"required"`
	SecretKey    string `json:"secret_key,omitempty" binding:"required"`
}

//GetAccessKeyReq get enterprise access key
//swagger:model GetAccessKeyReq
type GetAccessKeyReq struct {
	ProviderName string `form:"provider_name" binding:"required"`
}

//KubernetesClustersResponse list kclusters response
//swagger:model KubernetesClustersResponse
type KubernetesClustersResponse struct {
	Clusters []*v1alpha1.Cluster `json:"clusters"`
}

//AccessKeyResponse access key
//swagger:model AccessKeyResponse
type AccessKeyResponse struct {
	models.CloudAccessKey
}

//CreateKubernetesReq create kubernetes req
//swagger:model CreateKubernetesReq
type CreateKubernetesReq struct {
	Name               string   `json:"name" binding:"required"`
	WorkerResourceType string   `json:"resourceType"`
	WorkerNum          int      `json:"workerNum"`
	Provider           string   `json:"provider_name" binding:"required"`
	Region             string   `json:"region"`
	EIP                []string `json:"eip,omitempty"`
	// rke
	Nodes v1alpha1.NodeList `json:"nodes,omitempty"`
	// custom
	KubeConfig string `json:"kubeconfig,omitempty"`
}

//UpdateKubernetesReq update kubernetes req
//swagger:model UpdateKubernetesReq
type UpdateKubernetesReq struct {
	Provider           string            `json:"provider"`
	ClusterID          string            `json:"clusterID"`
	Nodes              v1alpha1.NodeList `json:"nodes,omitempty"`
	WorkerResourceType string            `json:"workerResourceType,omitempty"`
	WorkerNodeNum      int               `json:"workerNum,omitempty"`
	MasterNodeNum      int               `json:"masterNodeNum,omitempty"`
	ETCDNodeNum        int               `json:"etcdNodeNum,omitempty"`
	InstanceType       string            `json:"instanceType,omitempty"`
}

//CreateKubernetesRes create kubernetes res
//swagger:model CreateKubernetesRes
type CreateKubernetesRes struct {
	models.CreateKubernetesTask
}

//UpdateKubernetesRes create kubernetes res
//swagger:model UpdateKubernetesRes
type UpdateKubernetesRes struct {
	Task     *models.UpdateKubernetesTask `json:"task"`
	NodeList v1alpha1.NodeList            `json:"nodeList"`
}

//GetLastCreateKubernetesClusterTaskReq get last create kubernetes task
//swagger:model GetLastCreateKubernetesClusterTaskReq
type GetLastCreateKubernetesClusterTaskReq struct {
	ProviderName string `form:"provider_name" binding:"required"`
}

//DeleteKubernetesClusterReq delete cluster
//swagger:model DeleteKubernetesClusterReq
type DeleteKubernetesClusterReq struct {
	ProviderName string `form:"provider_name" binding:"required"`
}

//GetCreateKubernetesClusterTaskRes create kubernetes res
//swagger:model GetCreateKubernetesClusterTaskRes
type GetCreateKubernetesClusterTaskRes struct {
	models.CreateKubernetesTask
}

//GetTaskEventListReq get event list of task
//swagger:model GetTaskEventListReq
type GetTaskEventListReq struct {
	TaskID string `form:"taskID" binding:"required"`
}

//TaskEventListRes get event list of task
//swagger:model TaskEventListRes
type TaskEventListRes struct {
	Events []*models.TaskEvent `json:"events"`
}

//InitRainbondRegionReq init rainbond region
//swagger:model InitRainbondRegionReq
type InitRainbondRegionReq struct {
	Provider  string `json:"providerName" binding:"required"`
	ClusterID string `json:"clusterID" binding:"required"`
	Retry     bool   `json:"retry"`
}

//InitRainbondTaskRes init rainbond region response
//swagger:model InitRainbondTaskRes
type InitRainbondTaskRes struct {
	models.InitRainbondTask
}

//GetInitRainbondTaskReq get init rainbond task
//swagger:model GetInitRainbondTaskReq
type GetInitRainbondTaskReq struct {
	ProviderName string `form:"provider_name" binding:"required"`
}

// InitRainbondTaskListRes running init tasks
//swagger:model InitRainbondTaskListRes
type InitRainbondTaskListRes struct {
	Tasks []*models.InitRainbondTask `json:"tasks"`
}

// GetRegionConfigRes region configs
//swagger:model GetRegionConfigRes
type GetRegionConfigRes struct {
	Configs    map[string]string `json:"configs"`
	ConfigYaml string            `json:"configs_yaml"`
}

//GetRegionConfigReq get rainbond region config
//swagger:model GetRegionConfigReq
type GetRegionConfigReq struct {
	ProviderName string `form:"provider_name" binding:"required"`
}

//UpdateInitRainbondTaskStatusReq update init task status
//swagger:model UpdateInitRainbondTaskStatusReq
type UpdateInitRainbondTaskStatusReq struct {
	Status string `json:"status" binding:"required"`
}

//InitNodeCmdRes init node cmd
//swagger:model InitNodeCmdRes
type InitNodeCmdRes struct {
	Cmd string `json:"cmd"`
}

//GetLogContentRes create kubernetes cluster log
//swagger:model GetLogContentRes
type GetLogContentRes struct {
	Content string `json:"content"`
}

//GetKubeConfigRes get kubernetes cluster kubeconfig file
//swagger:model GetKubeConfigRes
type GetKubeConfigRes struct {
	Config string `json:"config"`
}
