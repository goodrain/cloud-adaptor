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

package usecase

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"

	"goodrain.com/cloud-adaptor/adaptor"
	"goodrain.com/cloud-adaptor/adaptor/custom"
	"goodrain.com/cloud-adaptor/adaptor/factory"
	"goodrain.com/cloud-adaptor/adaptor/rke"
	"goodrain.com/cloud-adaptor/adaptor/v1alpha1"
	"goodrain.com/cloud-adaptor/operator"
	"goodrain.com/cloud-adaptor/task"

	"goodrain.com/cloud-adaptor/api/cluster"
	"goodrain.com/cloud-adaptor/api/cluster/repository"
	"goodrain.com/cloud-adaptor/api/models"
	"goodrain.com/cloud-adaptor/api/nsqc/producer"
	"goodrain.com/cloud-adaptor/library/bcode"

	rainbondv1alpha1 "github.com/goodrain/rainbond-operator/api/v1alpha1"
	"goodrain.com/cloud-adaptor/util/md5util"
	"goodrain.com/cloud-adaptor/util/uuidutil"

	v1 "goodrain.com/cloud-adaptor/api/openapi/types/v1"
)

// ClusterUsecase cluster manage usecase
type ClusterUsecase struct {
	DB                        *gorm.DB                                `inject:""`
	CloudAccessKeyRepo        cluster.CloudAccesskeyRepository        `inject:""`
	CreateKubernetesTaskRepo  cluster.CreateKubernetesTaskRepository  `inject:""`
	InitRainbondTaskRepo      cluster.InitRainbondTaskRepository      `inject:""`
	UpdateKubernetesTaskRepo  cluster.UpdateKubernetesTaskRepository  `inject:""`
	TaskEventRepo             cluster.TaskEventRepository             `inject:""`
	TaskProducer              producer.TaskProducer                   `inject:""`
	RainbondClusterConfigRepo cluster.RainbondClusterConfigRepository `inject:""`
}

// NewClusterUsecase new cluster usecase
func NewClusterUsecase(taskProducer producer.TaskProducer) cluster.Usecase {
	return &ClusterUsecase{TaskProducer: taskProducer}
}

//ListKubernetesCluster list kubernetes cluster
func (c *ClusterUsecase) ListKubernetesCluster(eid string, re v1.ListKubernetesCluster) ([]*v1alpha1.Cluster, error) {
	var ad adaptor.RainbondClusterAdaptor
	var err error
	if re.ProviderName != "rke" && re.ProviderName != "custom" {
		accessKey, err := c.CloudAccessKeyRepo.GetByProviderAndEnterprise(re.ProviderName, eid)
		if err != nil {
			return nil, bcode.ErrorNotFoundAccessKey
		}
		ad, err = factory.GetCloudFactory().GetRainbondClusterAdaptor(re.ProviderName, accessKey.AccessKey, accessKey.SecretKey)
		if err != nil {
			return nil, bcode.ErrorProviderNotSupport
		}
	} else {
		ad, err = factory.GetCloudFactory().GetRainbondClusterAdaptor(re.ProviderName, "", "")
		if err != nil {
			return nil, bcode.ErrorProviderNotSupport
		}
	}
	clusters, err := ad.ClusterList(eid)
	if err != nil {
		if strings.Contains(err.Error(), "ErrorCode: SignatureDoesNotMatch") {
			return nil, bcode.ErrorAccessKeyNotMatch
		}
		if strings.Contains(err.Error(), "ErrorCode: InvalidAccessKeyId.NotFound") {
			return nil, bcode.ErrorNotFoundAccessKey
		}
		if strings.Contains(err.Error(), "Code: EntityNotExist.Role") {
			return nil, bcode.ErrorClusterRoleNotExist
		}
		logrus.Errorf("list cluster list failure %s", err.Error())
		return nil, bcode.ServerErr
	}
	return clusters, nil
}

//CreateKubernetesCluster create kubernetes cluster task
func (c *ClusterUsecase) CreateKubernetesCluster(eid string, req v1.CreateKubernetesReq) (*models.CreateKubernetesTask, error) {
	if c.TaskProducer == nil {
		logrus.Errorf("TaskProducer is nil")
		return nil, bcode.ServerErr
	}
	if req.Provider == "custom" {
		if err := custom.NewCustomClusterRepo(c.DB).Create(&models.CustomCluster{
			Name:         req.Name,
			EIP:          strings.Join(req.EIP, ","),
			KubeConfig:   req.KubeConfig,
			EnterpriseID: eid,
		}); err != nil {
			logrus.Errorf("create custom cluster failure %s", err.Error())
			return nil, bcode.ServerErr
		}
	}
	var accessKey *models.CloudAccessKey
	var err error
	if req.Provider != "rke" && req.Provider != "custom" {
		accessKey, err = c.CloudAccessKeyRepo.GetByProviderAndEnterprise(req.Provider, eid)
		if err != nil {
			return nil, bcode.ErrorNotFoundAccessKey
		}
	}
	newTask := &models.CreateKubernetesTask{
		Name:               req.Name,
		Provider:           req.Provider,
		WorkerResourceType: req.WorkerResourceType,
		WorkerNum:          req.WorkerNum,
		EnterpriseID:       eid,
		Region:             req.Region,
		TaskID:             uuidutil.NewUUID(),
	}
	if err := c.CreateKubernetesTaskRepo.Create(newTask); err != nil {
		logrus.Errorf("create kubernetes task failure %s", err.Error())
		return nil, bcode.ServerErr
	}
	//send task
	taskReq := task.KubernetesConfigMessage{
		EnterpriseID: eid,
		TaskID:       newTask.TaskID,
		KubernetesConfig: &v1alpha1.KubernetesClusterConfig{
			ClusterName:        newTask.Name,
			WorkerResourceType: newTask.WorkerResourceType,
			WorkerNodeNum:      newTask.WorkerNum,
			Provider:           newTask.Provider,
			Region:             newTask.Region,
			Nodes:              req.Nodes,
			EnterpriseID:       eid,
		}}
	if accessKey != nil {
		taskReq.KubernetesConfig.AccessKey = accessKey.AccessKey
		taskReq.KubernetesConfig.SecretKey = accessKey.SecretKey
	}
	if err := c.TaskProducer.SendCreateKuerbetesTask(taskReq); err != nil {
		logrus.Errorf("send create kubernetes task failure %s", err.Error())
	} else {
		if err := c.CreateKubernetesTaskRepo.UpdateStatus(eid, newTask.TaskID, "start"); err != nil {
			logrus.Errorf("update task status failure %s", err.Error())
		}
	}
	logrus.Infof("send create kubernetes task %s to queue", newTask.TaskID)
	return newTask, nil
}

//InitRainbondRegion init rainbond region
func (c *ClusterUsecase) InitRainbondRegion(eid string, req v1.InitRainbondRegionReq) (*models.InitRainbondTask, error) {
	oldTask, err := c.InitRainbondTaskRepo.GetTaskByClusterID(eid, req.Provider, req.ClusterID)
	if err != nil && err != gorm.ErrRecordNotFound {
		logrus.Errorf("query last init rainbond task failure %s", err.Error())
		return nil, bcode.ServerErr
	}
	if oldTask != nil && !req.Retry {
		return oldTask, bcode.ErrorLastTaskNotComplete
	}
	var accessKey *models.CloudAccessKey
	if req.Provider != "rke" && req.Provider != "custom" {
		accessKey, err = c.CloudAccessKeyRepo.GetByProviderAndEnterprise(req.Provider, eid)
		if err != nil {
			return nil, bcode.ErrorNotFoundAccessKey
		}
	}
	newTask := &models.InitRainbondTask{
		TaskID:       uuidutil.NewUUID(),
		Provider:     req.Provider,
		EnterpriseID: eid,
		ClusterID:    req.ClusterID,
	}

	if err := c.InitRainbondTaskRepo.Create(newTask); err != nil {
		logrus.Errorf("create init rainbond task failure %s", err.Error())
		return nil, bcode.ServerErr
	}
	initTask := task.InitRainbondConfigMessage{
		EnterpriseID: eid,
		TaskID:       newTask.TaskID,
		InitRainbondConfig: &task.InitRainbondConfig{
			EnterpriseID: eid,
			ClusterID:    newTask.ClusterID,
			Provider:     newTask.Provider,
		}}
	if accessKey != nil {
		initTask.InitRainbondConfig.AccessKey = accessKey.AccessKey
		initTask.InitRainbondConfig.SecretKey = accessKey.SecretKey
	}
	if err := c.TaskProducer.SendInitRainbondRegionTask(initTask); err != nil {
		logrus.Errorf("send init rainbond region task failure %s", err.Error())
	} else {
		if err := c.InitRainbondTaskRepo.UpdateStatus(eid, newTask.TaskID, "start"); err != nil {
			logrus.Errorf("update task status failure %s", err.Error())
		}
	}
	logrus.Infof("send init rainbond region task %s to queue", newTask.TaskID)
	return newTask, nil
}

//UpdateKubernetesCluster -
func (c *ClusterUsecase) UpdateKubernetesCluster(eid string, req v1.UpdateKubernetesReq) (*models.UpdateKubernetesTask, error) {
	if c.TaskProducer == nil {
		logrus.Errorf("TaskProducer is nil")
		return nil, bcode.ServerErr
	}
	if req.Provider != "rke" {
		return nil, bcode.ErrNotSupportUpdateKubernetes
	}
	newTask := &models.UpdateKubernetesTask{
		TaskID:       uuidutil.NewUUID(),
		Provider:     req.Provider,
		EnterpriseID: eid,
		ClusterID:    req.ClusterID,
		NodeNumber:   len(req.Nodes),
	}
	if err := c.UpdateKubernetesTaskRepo.Create(newTask); err != nil {
		logrus.Errorf("save update kubernetes task failure %s", err.Error())
		return nil, bcode.ServerErr
	}
	//send task
	taskReq := task.UpdateKubernetesConfigMessage{
		EnterpriseID: eid,
		TaskID:       newTask.TaskID,
		Config: &v1alpha1.ExpansionNode{
			Provider:     req.Provider,
			ClusterID:    req.ClusterID,
			Nodes:        req.Nodes,
			EnterpriseID: eid,
		}}
	if err := c.TaskProducer.SendUpdateKuerbetesTask(taskReq); err != nil {
		logrus.Errorf("send create kubernetes task failure %s", err.Error())
	} else {
		if err := c.CreateKubernetesTaskRepo.UpdateStatus(eid, newTask.TaskID, "start"); err != nil {
			logrus.Errorf("update task status failure %s", err.Error())
		}
	}
	logrus.Infof("send create kubernetes task %s to queue", newTask.TaskID)
	return newTask, nil
}

//GetInitRainbondTaskByClusterID get init rainbond task
func (c *ClusterUsecase) GetInitRainbondTaskByClusterID(eid, clusterID, providerName string) (*models.InitRainbondTask, error) {
	task, err := c.InitRainbondTaskRepo.GetTaskByClusterID(eid, providerName, clusterID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, bcode.NotFound
		}
		return nil, bcode.ServerErr
	}
	return task, nil
}

//GetUpdateKubernetesTaskByClusterID get update kubernetes task
func (c *ClusterUsecase) GetUpdateKubernetesTaskByClusterID(eid, clusterID, providerName string) (*models.UpdateKubernetesTask, error) {
	task, err := c.UpdateKubernetesTaskRepo.GetTaskByClusterID(eid, providerName, clusterID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, bcode.ServerErr
	}
	return task, nil
}

//GetRKENodeList get rke kubernetes node list
func (c *ClusterUsecase) GetRKENodeList(eid, clusterID string) (v1alpha1.NodeList, error) {
	cluster, err := rke.NewRKEClusterRepo(c.DB).GetCluster(eid, clusterID)
	if err != nil {
		return nil, bcode.NotFound
	}
	var nodes v1alpha1.NodeList
	if err := json.Unmarshal([]byte(cluster.NodeList), &nodes); err != nil {
		return nil, err
	}
	return nodes, nil
}

//AddAccessKey add accesskey info to enterprise
func (c *ClusterUsecase) AddAccessKey(eid string, key v1.AddAccessKey) (*models.CloudAccessKey, error) {
	ack, err := c.GetByProviderAndEnterprise(key.ProviderName, eid)
	if err != nil && err != bcode.ErrorNotSetAccessKey {
		return nil, err
	}
	if ack != nil && key.AccessKey == ack.AccessKey && key.SecretKey == md5util.Md5Crypt(ack.SecretKey, ack.EnterpriseID) {
		return ack, nil
	}

	ck := &models.CloudAccessKey{
		EnterpriseID: eid,
		ProviderName: key.ProviderName,
		AccessKey:    key.AccessKey,
		SecretKey:    key.SecretKey,
	}
	if err := c.CloudAccessKeyRepo.Create(ck); err != nil {
		return nil, err
	}
	return ck, nil
}

//GetByProviderAndEnterprise get by eid
func (c *ClusterUsecase) GetByProviderAndEnterprise(providerName, eid string) (*models.CloudAccessKey, error) {
	key, err := c.CloudAccessKeyRepo.GetByProviderAndEnterprise(providerName, eid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, bcode.ErrorNotSetAccessKey
		}
		return nil, bcode.ServerErr
	}
	return key, nil
}

//CreateTaskEvent create task event
func (c *ClusterUsecase) CreateTaskEvent(em *v1.EventMessage) (*models.TaskEvent, error) {
	if em.Message == nil {
		return nil, fmt.Errorf("message is nil")
	}
	ctx := c.DB.Begin()
	ent := &models.TaskEvent{
		TaskID:       em.TaskID,
		EnterpriseID: em.EnterpriseID,
		Status:       em.Message.Status,
		StepType:     em.Message.StepType,
		Message:      em.Message.Message,
	}
	if err := repository.NewTaskEventRepo(ctx).Create(ent); err != nil {
		ctx.Rollback()
		return nil, err
	}

	if (em.Message.StepType == "CreateCluster" || em.Message.StepType == "InstallKubernetes") && em.Message.Status == "success" {
		if ckErr := repository.NewCreateKubernetesTaskRepo(ctx).UpdateStatus(em.EnterpriseID, em.TaskID, "complete"); ckErr != nil && ckErr != gorm.ErrRecordNotFound {
			ctx.Rollback()
			return nil, ckErr
		}
		logrus.Infof("set create kubernetes task %s status is complete", em.TaskID)
	}
	if em.Message.StepType == "InitRainbondRegion" && em.Message.Status == "success" {
		if err := repository.NewInitRainbondRegionTaskRepo(ctx).UpdateStatus(em.EnterpriseID, em.TaskID, "inited"); err != nil && err != gorm.ErrRecordNotFound {
			ctx.Rollback()
			return nil, err
		}
		logrus.Infof("set init task %s status is inited", em.TaskID)
	}
	if em.Message.StepType == "UpdateKubernetes" && em.Message.Status == "success" {
		if err := repository.NewUpdateKubernetesTaskRepo(ctx).UpdateStatus(em.EnterpriseID, em.TaskID, "complete"); err != nil && err != gorm.ErrRecordNotFound {
			ctx.Rollback()
			return nil, err
		}
		logrus.Infof("set init task %s status is inited", em.TaskID)
	}
	if em.Message.Status == "failure" {
		if initErr := repository.NewInitRainbondRegionTaskRepo(ctx).UpdateStatus(em.EnterpriseID, em.TaskID, "complete"); initErr != nil && initErr != gorm.ErrRecordNotFound {
			ctx.Rollback()
			return nil, initErr
		}

		if ckErr := repository.NewCreateKubernetesTaskRepo(ctx).UpdateStatus(em.EnterpriseID, em.TaskID, "complete"); ckErr != nil && ckErr != gorm.ErrRecordNotFound {
			ctx.Rollback()
			return nil, ckErr
		}
	}

	if err := ctx.Commit().Error; err != nil {
		ctx.Rollback()
		return nil, err
	}
	logrus.Infof("save task %s event %s status %s to db", em.TaskID, em.Message.StepType, em.Message.Status)
	return ent, nil
}

//ListTaskEvent list task event list
func (c *ClusterUsecase) ListTaskEvent(eid, taskID string) ([]*models.TaskEvent, error) {
	events, err := c.TaskEventRepo.ListEvent(eid, taskID)
	if err != nil {
		return nil, err
	}
	for i := range events {
		event := events[i]
		if (event.StepType == "CreateCluster" || event.StepType == "InstallKubernetes") && event.Status == "success" {
			if ckErr := c.CreateKubernetesTaskRepo.UpdateStatus(eid, event.TaskID, "complete"); ckErr != nil && ckErr != gorm.ErrRecordNotFound {
				logrus.Errorf("set create kubernetes task %s status failure %s", event.TaskID, err.Error())
			}
			logrus.Infof("set create kubernetes task %s status is complete", event.TaskID)
		}
		if event.StepType == "InitRainbondRegion" && event.Status == "success" {
			if err := c.InitRainbondTaskRepo.UpdateStatus(eid, event.TaskID, "inited"); err != nil && err != gorm.ErrRecordNotFound {
				logrus.Errorf("set init rainbond task %s status failure %s", event.TaskID, err.Error())
			}
			logrus.Infof("set init task %s status is inited", event.TaskID)
		}
		if event.StepType == "UpdateKubernetes" && event.Status == "success" {
			if err := c.UpdateKubernetesTaskRepo.UpdateStatus(eid, event.TaskID, "complete"); err != nil && err != gorm.ErrRecordNotFound {
				logrus.Errorf("set init rainbond task %s status failure %s", event.TaskID, err.Error())
			}
			logrus.Infof("set init task %s status is inited", event.TaskID)
		}
		if event.Status == "failure" {
			if initErr := c.InitRainbondTaskRepo.UpdateStatus(eid, event.TaskID, "complete"); initErr != nil && initErr != gorm.ErrRecordNotFound {
				logrus.Errorf("set init rainbond task %s status failure %s", event.TaskID, err.Error())
			}

			if ckErr := c.CreateKubernetesTaskRepo.UpdateStatus(eid, event.TaskID, "complete"); ckErr != nil && ckErr != gorm.ErrRecordNotFound {
				logrus.Errorf("set create kubernetes task %s status failure %s", event.TaskID, err.Error())
			}

			if ckErr := c.UpdateKubernetesTaskRepo.UpdateStatus(eid, event.TaskID, "complete"); ckErr != nil && ckErr != gorm.ErrRecordNotFound {
				logrus.Errorf("set update kubernetes task %s status failure %s", event.TaskID, err.Error())
			}
		}
	}
	return events, nil
}

//GetLastCreateKubernetesTask get last create kubernetes task
func (c *ClusterUsecase) GetLastCreateKubernetesTask(eid, providerName string) (*models.CreateKubernetesTask, error) {
	task, err := c.CreateKubernetesTaskRepo.GetLastTask(eid, providerName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return task, nil
}

//GetCreateKubernetesTask get task
func (c *ClusterUsecase) GetCreateKubernetesTask(eid, taskID string) (*models.CreateKubernetesTask, error) {
	task, err := c.CreateKubernetesTaskRepo.GetTask(eid, taskID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, bcode.NotFound
		}
		return nil, err
	}
	return task, err
}

//GetTaskRunningLists get runinig tasks
func (c *ClusterUsecase) GetTaskRunningLists(eid string) ([]*models.InitRainbondTask, error) {
	tasks, err := c.InitRainbondTaskRepo.GetTaskRunningLists(eid)
	if err != nil {
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, nil
			}
			return nil, err
		}
	}
	return tasks, nil
}

//GetKubeConfig get kube config file
func (c *ClusterUsecase) GetKubeConfig(eid, clusterID, providerName string) (string, error) {
	var ad adaptor.RainbondClusterAdaptor
	var err error
	if providerName != "rke" && providerName != "custom" {
		accessKey, err := c.CloudAccessKeyRepo.GetByProviderAndEnterprise(providerName, eid)
		if err != nil {
			return "", bcode.ErrorNotFoundAccessKey
		}
		ad, err = factory.GetCloudFactory().GetRainbondClusterAdaptor(providerName, accessKey.AccessKey, accessKey.SecretKey)
		if err != nil {
			return "", bcode.ErrorProviderNotSupport
		}
	} else {
		ad, err = factory.GetCloudFactory().GetRainbondClusterAdaptor(providerName, "", "")
		if err != nil {
			return "", bcode.ErrorProviderNotSupport
		}
	}
	kube, err := ad.GetKubeConfig(eid, clusterID)
	if err != nil {
		return "", err
	}
	return kube.Config, nil
}

//GetRegionConfig get region config
func (c *ClusterUsecase) GetRegionConfig(eid, clusterID, providerName string) (map[string]string, error) {
	var ad adaptor.RainbondClusterAdaptor
	var err error
	if providerName != "rke" && providerName != "custom" {
		accessKey, err := c.CloudAccessKeyRepo.GetByProviderAndEnterprise(providerName, eid)
		if err != nil {
			return nil, bcode.ErrorNotFoundAccessKey
		}
		ad, err = factory.GetCloudFactory().GetRainbondClusterAdaptor(providerName, accessKey.AccessKey, accessKey.SecretKey)
		if err != nil {
			return nil, bcode.ErrorProviderNotSupport
		}
	} else {
		ad, err = factory.GetCloudFactory().GetRainbondClusterAdaptor(providerName, "", "")
		if err != nil {
			return nil, bcode.ErrorProviderNotSupport
		}
	}
	kubeConfig, err := ad.GetKubeConfig(eid, clusterID)
	if err != nil {
		return nil, bcode.ErrorKubeAPI
	}
	rri := operator.NewRainbondRegionInit(*kubeConfig)
	status, err := rri.GetRainbondRegionStatus(clusterID)
	if err != nil {
		logrus.Errorf("get rainbond region status failure %s", err.Error())
		return nil, bcode.ErrorGetRegionStatus
	}
	if status.RegionConfig != nil {
		configMap := status.RegionConfig
		regionConfig := map[string]string{
			"client.pem":          string(configMap.BinaryData["client.pem"]),
			"client.key.pem":      string(configMap.BinaryData["client.key.pem"]),
			"ca.pem":              string(configMap.BinaryData["ca.pem"]),
			"apiAddress":          configMap.Data["apiAddress"],
			"websocketAddress":    configMap.Data["websocketAddress"],
			"defaultDomainSuffix": configMap.Data["defaultDomainSuffix"],
			"defaultTCPHost":      configMap.Data["defaultTCPHost"],
		}
		return regionConfig, nil
	}
	return nil, nil
}

//UpdateInitRainbondTaskStatus update init rainbond task status
func (c *ClusterUsecase) UpdateInitRainbondTaskStatus(eid, taskID, status string) (*models.InitRainbondTask, error) {
	if err := c.InitRainbondTaskRepo.UpdateStatus(eid, taskID, status); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, bcode.NotFound
		}
		return nil, err
	}
	task, err := c.InitRainbondTaskRepo.GetTask(eid, taskID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, bcode.NotFound
		}
		return nil, err
	}
	return task, nil
}

//DeleteKubernetesCluster delete provider
func (c *ClusterUsecase) DeleteKubernetesCluster(eid, clusterID, providerName string) error {
	var ad adaptor.RainbondClusterAdaptor
	var err error
	if providerName != "rke" && providerName != "custom" {
		accessKey, err := c.CloudAccessKeyRepo.GetByProviderAndEnterprise(providerName, eid)
		if err != nil {
			return bcode.ErrorNotFoundAccessKey
		}
		ad, err = factory.GetCloudFactory().GetRainbondClusterAdaptor(providerName, accessKey.AccessKey, accessKey.SecretKey)
		if err != nil {
			return bcode.ErrorProviderNotSupport
		}
	} else {
		ad, err = factory.GetCloudFactory().GetRainbondClusterAdaptor(providerName, "", "")
		if err != nil {
			return bcode.ErrorProviderNotSupport
		}
	}
	return ad.DeleteCluster(eid, clusterID)
}

//GetCluster get cluster
func (c *ClusterUsecase) GetCluster(providerName, eid, clusterID string) (*v1alpha1.Cluster, error) {
	var ad adaptor.RainbondClusterAdaptor
	var err error
	if providerName != "rke" && providerName != "custom" {
		accessKey, err := c.CloudAccessKeyRepo.GetByProviderAndEnterprise(providerName, eid)
		if err != nil {
			return nil, bcode.ErrorNotFoundAccessKey
		}
		ad, err = factory.GetCloudFactory().GetRainbondClusterAdaptor(providerName, accessKey.AccessKey, accessKey.SecretKey)
		if err != nil {
			return nil, bcode.ErrorProviderNotSupport
		}
	} else {
		ad, err = factory.GetCloudFactory().GetRainbondClusterAdaptor(providerName, "", "")
		if err != nil {
			return nil, bcode.ErrorProviderNotSupport
		}
	}
	return ad.DescribeCluster(eid, clusterID)
}

//InstallCluster install cluster
func (c *ClusterUsecase) InstallCluster(eid, clusterID string) (*models.CreateKubernetesTask, error) {
	if c.TaskProducer == nil {
		logrus.Errorf("TaskProducer is nil")
		return nil, bcode.ServerErr
	}
	cluster, err := rke.NewRKEClusterRepo(c.DB).GetCluster(eid, clusterID)
	if err != nil {
		return nil, err
	}
	newTask := &models.CreateKubernetesTask{
		Name:         cluster.Name,
		Provider:     "rke",
		EnterpriseID: eid,
		TaskID:       uuidutil.NewUUID(),
	}
	if err := c.CreateKubernetesTaskRepo.Create(newTask); err != nil {
		logrus.Errorf("create kubernetes task failure %s", err.Error())
		return nil, bcode.ServerErr
	}
	var nodes v1alpha1.NodeList
	json.Unmarshal([]byte(cluster.NodeList), &nodes)
	//send task
	taskReq := task.KubernetesConfigMessage{
		EnterpriseID: eid,
		TaskID:       newTask.TaskID,
		KubernetesConfig: &v1alpha1.KubernetesClusterConfig{
			ClusterName:  newTask.Name,
			Provider:     newTask.Provider,
			Nodes:        nodes,
			EnterpriseID: eid,
		}}
	if err := c.TaskProducer.SendCreateKuerbetesTask(taskReq); err != nil {
		logrus.Errorf("send create kubernetes task failure %s", err.Error())
	} else {
		if err := c.CreateKubernetesTaskRepo.UpdateStatus(eid, newTask.TaskID, "start"); err != nil {
			logrus.Errorf("update task status failure %s", err.Error())
		}
	}
	logrus.Infof("send create kubernetes task %s to queue", newTask.TaskID)
	return newTask, nil
}

//SetRainbondClusterConfig set rainbond cluster config
func (c *ClusterUsecase) SetRainbondClusterConfig(eid, clusterID, config string) error {
	var rbcc rainbondv1alpha1.RainbondCluster
	if err := yaml.Unmarshal([]byte(config), &rbcc); err != nil {
		logrus.Errorf("unmarshal rainbond config failure %s", err.Error())
		return bcode.ErrConfigInvalid
	}
	return c.RainbondClusterConfigRepo.Create(
		&models.RainbondClusterConfig{
			ClusterID:    clusterID,
			Config:       config,
			EnterpriseID: eid,
		})
}

//GetRainbondClusterConfig get rainbond cluster config
func (c *ClusterUsecase) GetRainbondClusterConfig(eid, clusterID string) (*rainbondv1alpha1.RainbondCluster, error) {
	rcc, _ := c.RainbondClusterConfigRepo.Get(clusterID)
	var rbcc rainbondv1alpha1.RainbondCluster
	rbcc.Name = "rainbondcluster"
	rbcc.Spec.EtcdConfig = &rainbondv1alpha1.EtcdConfig{}
	rbcc.Spec.ImageHub = &rainbondv1alpha1.ImageHub{}
	rbcc.Spec.NodesForGateway = []*rainbondv1alpha1.K8sNode{}
	rbcc.Spec.NodesForChaos = []*rainbondv1alpha1.K8sNode{}
	rbcc.Spec.SuffixHTTPHost = ""
	rbcc.Spec.GatewayIngressIPs = []string{}
	if rcc != nil {
		if err := yaml.Unmarshal([]byte(rcc.Config), &rbcc); err != nil {
			logrus.Errorf("unmarshal rainbond config failure %s", err.Error())
		}
	}
	return &rbcc, nil
}

//UninstallRainbondRegion uninstall rainbond region
func (c *ClusterUsecase) UninstallRainbondRegion(eid, clusterID, provider string) error {
	if os.Getenv("DISABLE_UNINSTALL_REGION") == "true" {
		logrus.Info("uninstall rainbond region is disable")
		return nil
	}
	var ad adaptor.RainbondClusterAdaptor
	var err error
	if provider != "rke" && provider != "custom" {
		accessKey, err := c.CloudAccessKeyRepo.GetByProviderAndEnterprise(provider, eid)
		if err != nil {
			return bcode.ErrorNotFoundAccessKey
		}
		ad, err = factory.GetCloudFactory().GetRainbondClusterAdaptor(provider, accessKey.AccessKey, accessKey.SecretKey)
		if err != nil {
			return bcode.ErrorProviderNotSupport
		}
	} else {
		ad, err = factory.GetCloudFactory().GetRainbondClusterAdaptor(provider, "", "")
		if err != nil {
			return bcode.ErrorProviderNotSupport
		}
	}
	kubeconfig, err := ad.GetKubeConfig(eid, clusterID)
	if err != nil {
		return err
	}
	rri := operator.NewRainbondRegionInit(*kubeconfig)
	go func() {
		logrus.Infof("start uninstall cluster %s by provider %s", clusterID, provider)
		if err := rri.UninstallRegion(clusterID); err != nil {
			logrus.Errorf("uninstall region %s failure %s", err.Error())
		}
		if err := c.InitRainbondTaskRepo.DeleteTask(eid, provider, clusterID); err != nil {
			logrus.Errorf("delete region init task failure %s", err.Error())
		}
		logrus.Infof("complete uninstall cluster %s by provider %s", clusterID, provider)
	}()
	return nil
}
