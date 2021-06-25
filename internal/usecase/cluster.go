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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	rainbondv1alpha1 "github.com/goodrain/rainbond-operator/api/v1alpha1"
	"github.com/pkg/errors"
	v3 "github.com/rancher/rke/types"
	"github.com/sirupsen/logrus"
	v1 "goodrain.com/cloud-adaptor/api/cloud-adaptor/v1"
	"goodrain.com/cloud-adaptor/internal/adaptor"
	"goodrain.com/cloud-adaptor/internal/adaptor/custom"
	"goodrain.com/cloud-adaptor/internal/adaptor/factory"
	"goodrain.com/cloud-adaptor/internal/adaptor/v1alpha1"
	"goodrain.com/cloud-adaptor/internal/model"
	"goodrain.com/cloud-adaptor/internal/nsqc/producer"
	"goodrain.com/cloud-adaptor/internal/operator"
	"goodrain.com/cloud-adaptor/internal/repo"
	"goodrain.com/cloud-adaptor/internal/types"
	"goodrain.com/cloud-adaptor/pkg/bcode"
	"goodrain.com/cloud-adaptor/pkg/util/md5util"
	"goodrain.com/cloud-adaptor/pkg/util/uuidutil"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
)

// ClusterUsecase cluster manage usecase
type ClusterUsecase struct {
	DB                        *gorm.DB
	TaskProducer              producer.TaskProducer
	CloudAccessKeyRepo        repo.CloudAccesskeyRepository
	CreateKubernetesTaskRepo  repo.CreateKubernetesTaskRepository
	InitRainbondTaskRepo      repo.InitRainbondTaskRepository
	UpdateKubernetesTaskRepo  repo.UpdateKubernetesTaskRepository
	TaskEventRepo             repo.TaskEventRepository
	RainbondClusterConfigRepo repo.RainbondClusterConfigRepository
	rkeClusterRepo            repo.RKEClusterRepository
}

// NewClusterUsecase new cluster usecase
func NewClusterUsecase(db *gorm.DB,
	taskProducer producer.TaskProducer,
	cloudAccessKeyRepo repo.CloudAccesskeyRepository,
	CreateKubernetesTaskRepo repo.CreateKubernetesTaskRepository,
	InitRainbondTaskRepo repo.InitRainbondTaskRepository,
	UpdateKubernetesTaskRepo repo.UpdateKubernetesTaskRepository,
	TaskEventRepo repo.TaskEventRepository,
	RainbondClusterConfigRepo repo.RainbondClusterConfigRepository,
	rkeClusterRepo repo.RKEClusterRepository,
) *ClusterUsecase {
	return &ClusterUsecase{
		DB:                        db,
		TaskProducer:              taskProducer,
		CloudAccessKeyRepo:        cloudAccessKeyRepo,
		CreateKubernetesTaskRepo:  CreateKubernetesTaskRepo,
		InitRainbondTaskRepo:      InitRainbondTaskRepo,
		UpdateKubernetesTaskRepo:  UpdateKubernetesTaskRepo,
		TaskEventRepo:             TaskEventRepo,
		RainbondClusterConfigRepo: RainbondClusterConfigRepo,
		rkeClusterRepo:            rkeClusterRepo,
	}
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
func (c *ClusterUsecase) CreateKubernetesCluster(eid string, req v1.CreateKubernetesReq) (*model.CreateKubernetesTask, error) {
	if c.TaskProducer == nil {
		return nil, errors.New("TaskProducer is nil")
	}
	if req.Provider == "custom" {
		if err := custom.NewCustomClusterRepo(c.DB).Create(&model.CustomCluster{
			Name:         req.Name,
			EIP:          strings.Join(req.EIP, ","),
			KubeConfig:   req.KubeConfig,
			EnterpriseID: eid,
		}); err != nil {
			return nil, errors.Wrap(err, "create custom cluster")
		}
	}
	var accessKey *model.CloudAccessKey
	var err error
	if req.Provider != "rke" && req.Provider != "custom" {
		accessKey, err = c.CloudAccessKeyRepo.GetByProviderAndEnterprise(req.Provider, eid)
		if err != nil {
			return nil, bcode.ErrorNotFoundAccessKey
		}
	}
	newTask := &model.CreateKubernetesTask{
		Name:               req.Name,
		Provider:           req.Provider,
		WorkerResourceType: req.WorkerResourceType,
		WorkerNum:          req.WorkerNum,
		EnterpriseID:       eid,
		Region:             req.Region,
		TaskID:             uuidutil.NewUUID(),
	}
	if err := c.CreateKubernetesTaskRepo.Create(newTask); err != nil {
		return nil, errors.Wrap(err, "create kubernetes task")
	}
	//send task
	taskReq := types.KubernetesConfigMessage{
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
func (c *ClusterUsecase) InitRainbondRegion(eid string, req v1.InitRainbondRegionReq) (*model.InitRainbondTask, error) {
	oldTask, err := c.InitRainbondTaskRepo.GetTaskByClusterID(eid, req.Provider, req.ClusterID)
	if err != nil && err != gorm.ErrRecordNotFound {
		logrus.Errorf("query last init rainbond task failure %s", err.Error())
		return nil, bcode.ServerErr
	}
	if oldTask != nil && !req.Retry {
		return oldTask, bcode.ErrorLastTaskNotComplete
	}
	var accessKey *model.CloudAccessKey
	if req.Provider != "rke" && req.Provider != "custom" {
		accessKey, err = c.CloudAccessKeyRepo.GetByProviderAndEnterprise(req.Provider, eid)
		if err != nil {
			return nil, bcode.ErrorNotFoundAccessKey
		}
	}
	newTask := &model.InitRainbondTask{
		TaskID:       uuidutil.NewUUID(),
		Provider:     req.Provider,
		EnterpriseID: eid,
		ClusterID:    req.ClusterID,
	}

	if err := c.InitRainbondTaskRepo.Create(newTask); err != nil {
		logrus.Errorf("create init rainbond task failure %s", err.Error())
		return nil, bcode.ServerErr
	}
	initTask := types.InitRainbondConfigMessage{
		EnterpriseID: eid,
		TaskID:       newTask.TaskID,
		InitRainbondConfig: &types.InitRainbondConfig{
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
func (c *ClusterUsecase) UpdateKubernetesCluster(eid string, req v1.UpdateKubernetesReq) (*v1.UpdateKubernetesTask, error) {
	if c.TaskProducer == nil {
		logrus.Errorf("TaskProducer is nil")
		return nil, bcode.ServerErr
	}
	if req.Provider != "rke" {
		return nil, bcode.ErrNotSupportUpdateKubernetes
	}
	newTask := &model.UpdateKubernetesTask{
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
	taskReq := types.UpdateKubernetesConfigMessage{
		EnterpriseID: eid,
		TaskID:       newTask.TaskID,
		Config: &v1alpha1.ExpansionNode{
			Provider:     req.Provider,
			ClusterID:    req.ClusterID,
			Nodes:        req.Nodes,
			EnterpriseID: eid,
			RKEConfig:    req.RKEConfig,
		}}
	if err := c.TaskProducer.SendUpdateKuerbetesTask(taskReq); err != nil {
		logrus.Errorf("send create kubernetes task failure %s", err.Error())
	} else {
		if err := c.CreateKubernetesTaskRepo.UpdateStatus(eid, newTask.TaskID, "start"); err != nil {
			logrus.Errorf("update task status failure %s", err.Error())
		}
	}
	logrus.Infof("send create kubernetes task %s to queue", newTask.TaskID)
	return &v1.UpdateKubernetesTask{
		TaskID:       newTask.TaskID,
		Provider:     newTask.Provider,
		EnterpriseID: newTask.EnterpriseID,
		ClusterID:    newTask.ClusterID,
		NodeNumber:   newTask.NodeNumber,
	}, nil
}

//GetInitRainbondTaskByClusterID get init rainbond task
func (c *ClusterUsecase) GetInitRainbondTaskByClusterID(eid, clusterID, providerName string) (*model.InitRainbondTask, error) {
	task, err := c.InitRainbondTaskRepo.GetTaskByClusterID(eid, providerName, clusterID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, bcode.NotFound
		}
		return nil, bcode.ServerErr
	}
	return task, nil
}

//GetUpdateKubernetesTask get update kubernetes task
func (c *ClusterUsecase) GetUpdateKubernetesTask(eid, clusterID, providerName string) (*v1.UpdateKubernetesRes, error) {
	task, err := c.getUpdateKubernetesTask(eid, clusterID, providerName)
	if err != nil {
		return nil, err
	}

	var re v1.UpdateKubernetesRes
	re.Task = task
	if providerName == "rke" {
		cluster, err := c.rkeClusterRepo.GetCluster(eid, clusterID)
		if err != nil {
			return nil, err
		}

		nodeList, err := c.GetRKENodeList(eid, clusterID)
		if err != nil {
			return nil, err
		}
		re.NodeList = nodeList

		rkeConfig, err := c.getRKEConfig(eid, cluster)
		if err != nil {
			return nil, err
		}

		rkeConfigBytes, err := yaml.Marshal(rkeConfig)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		re.RKEConfig = base64.StdEncoding.EncodeToString(rkeConfigBytes)
	}

	return &re, nil
}

func (c *ClusterUsecase) getUpdateKubernetesTask(eid, clusterID, providerName string) (*model.UpdateKubernetesTask, error) {
	task, err := c.UpdateKubernetesTaskRepo.GetTaskByClusterID(eid, providerName, clusterID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return task, nil
}

func (c *ClusterUsecase) getRKEConfig(eid string, cluster *model.RKECluster) (*v3.RancherKubernetesEngineConfig, error) {
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		configDir = "/tmp"
	}
	clusterYMLPath := fmt.Sprintf("%s/enterprise/%s/rke/%s", configDir, cluster.EnterpriseID, cluster.Name)
	oldclusterYMLPath := fmt.Sprintf("%s/rke/%s/cluster.yml", configDir, cluster.Name)

	bytes, err := ioutil.ReadFile(clusterYMLPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "read rke config file")
		}
		bytes, err = ioutil.ReadFile(oldclusterYMLPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, errors.Wrap(err, "read old rke config file")
			}
			return nil, nil
		}
	}

	var rkeConfig v3.RancherKubernetesEngineConfig
	if err = yaml.Unmarshal(bytes, &rkeConfig); err != nil {
		return nil, errors.WithStack(bcode.ErrIncorrectRKEConfig)
	}

	return &rkeConfig, nil
}

//GetRKENodeList get rke kubernetes node list
func (c *ClusterUsecase) GetRKENodeList(eid, clusterID string) (v1alpha1.NodeList, error) {
	cluster, err := repo.NewRKEClusterRepo(c.DB).GetCluster(eid, clusterID)
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
func (c *ClusterUsecase) AddAccessKey(eid string, key v1.AddAccessKey) (*model.CloudAccessKey, error) {
	ack, err := c.GetByProviderAndEnterprise(key.ProviderName, eid)
	if err != nil && err != bcode.ErrorNotSetAccessKey {
		return nil, err
	}
	if ack != nil && key.AccessKey == ack.AccessKey && key.SecretKey == md5util.Md5Crypt(ack.SecretKey, ack.EnterpriseID) {
		return ack, nil
	}

	ck := &model.CloudAccessKey{
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
func (c *ClusterUsecase) GetByProviderAndEnterprise(providerName, eid string) (*model.CloudAccessKey, error) {
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
func (c *ClusterUsecase) CreateTaskEvent(em *v1.EventMessage) (*model.TaskEvent, error) {
	if em.Message == nil {
		return nil, fmt.Errorf("message is nil")
	}
	ctx := c.DB.Begin()
	ent := &model.TaskEvent{
		TaskID:       em.TaskID,
		EnterpriseID: em.EnterpriseID,
		Status:       em.Message.Status,
		StepType:     em.Message.StepType,
		Message:      em.Message.Message,
	}

	if err := c.TaskEventRepo.Transaction(ctx).Create(ent); err != nil {
		ctx.Rollback()
		return nil, err
	}

	createKubernetesTaskRepo := c.CreateKubernetesTaskRepo.Transaction(ctx)
	if (em.Message.StepType == "CreateCluster" || em.Message.StepType == "InstallKubernetes") && em.Message.Status == "success" {
		if ckErr := createKubernetesTaskRepo.UpdateStatus(em.EnterpriseID, em.TaskID, "complete"); ckErr != nil && ckErr != gorm.ErrRecordNotFound {
			ctx.Rollback()
			return nil, ckErr
		}
		logrus.Infof("set create kubernetes task %s status is complete", em.TaskID)
	}
	initRainbondTaskRepo := c.InitRainbondTaskRepo.Transaction(ctx)
	if em.Message.StepType == "InitRainbondRegion" && em.Message.Status == "success" {
		if err := initRainbondTaskRepo.UpdateStatus(em.EnterpriseID, em.TaskID, "inited"); err != nil && err != gorm.ErrRecordNotFound {
			ctx.Rollback()
			return nil, err
		}
		logrus.Infof("set init task %s status is inited", em.TaskID)
	}
	if em.Message.StepType == "UpdateKubernetes" && em.Message.Status == "success" {
		if err := c.UpdateKubernetesTaskRepo.Transaction(ctx).UpdateStatus(em.EnterpriseID, em.TaskID, "complete"); err != nil && err != gorm.ErrRecordNotFound {
			ctx.Rollback()
			return nil, err
		}
		logrus.Infof("set init task %s status is inited", em.TaskID)
	}
	if em.Message.Status == "failure" {
		if initErr := initRainbondTaskRepo.UpdateStatus(em.EnterpriseID, em.TaskID, "complete"); initErr != nil && initErr != gorm.ErrRecordNotFound {
			ctx.Rollback()
			return nil, initErr
		}

		if ckErr := createKubernetesTaskRepo.UpdateStatus(em.EnterpriseID, em.TaskID, "complete"); ckErr != nil && ckErr != gorm.ErrRecordNotFound {
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
func (c *ClusterUsecase) ListTaskEvent(eid, taskID string) ([]*model.TaskEvent, error) {
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
func (c *ClusterUsecase) GetLastCreateKubernetesTask(eid, providerName string) (*model.CreateKubernetesTask, error) {
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
func (c *ClusterUsecase) GetCreateKubernetesTask(eid, taskID string) (*model.CreateKubernetesTask, error) {
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
func (c *ClusterUsecase) GetTaskRunningLists(eid string) ([]*model.InitRainbondTask, error) {
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
	rri := operator.NewRainbondRegionInit(*kubeConfig, c.RainbondClusterConfigRepo)
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
func (c *ClusterUsecase) UpdateInitRainbondTaskStatus(eid, taskID, status string) (*model.InitRainbondTask, error) {
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
func (c *ClusterUsecase) InstallCluster(eid, clusterID string) (*model.CreateKubernetesTask, error) {
	if c.TaskProducer == nil {
		logrus.Errorf("TaskProducer is nil")
		return nil, bcode.ServerErr
	}
	cluster, err := repo.NewRKEClusterRepo(c.DB).GetCluster(eid, clusterID)
	if err != nil {
		return nil, err
	}
	newTask := &model.CreateKubernetesTask{
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
	taskReq := types.KubernetesConfigMessage{
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
		&model.RainbondClusterConfig{
			ClusterID:    clusterID,
			Config:       config,
			EnterpriseID: eid,
		})
}

//GetRainbondClusterConfig get rainbond cluster config
func (c *ClusterUsecase) GetRainbondClusterConfig(eid, clusterID string) (*rainbondv1alpha1.RainbondCluster, string) {
	rcc, _ := c.RainbondClusterConfigRepo.Get(clusterID)
	if rcc != nil {
		var rbcc rainbondv1alpha1.RainbondCluster
		if err := yaml.Unmarshal([]byte(rcc.Config), &rbcc); err != nil {
			logrus.Errorf("unmarshal rainbond config failure %s", err.Error())
			return nil, rcc.Config
		}
		return &rbcc, rcc.Config
	}
	return nil, ""
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
	rri := operator.NewRainbondRegionInit(*kubeconfig, c.RainbondClusterConfigRepo)
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
