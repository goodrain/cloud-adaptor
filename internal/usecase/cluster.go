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
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/devfeel/mapper"
	rainbondv1alpha1 "github.com/goodrain/rainbond-operator/api/v1alpha1"
	"github.com/goodrain/rainbond-operator/util/rbdutil"
	"github.com/pkg/errors"
	v3 "github.com/rancher/rke/types"
	"github.com/sirupsen/logrus"
	v1 "goodrain.com/cloud-adaptor/api/cloud-adaptor/v1"
	"goodrain.com/cloud-adaptor/cmd/cloud-adaptor/config"
	"goodrain.com/cloud-adaptor/internal/adaptor"
	"goodrain.com/cloud-adaptor/internal/adaptor/factory"
	"goodrain.com/cloud-adaptor/internal/adaptor/v1alpha1"
	"goodrain.com/cloud-adaptor/internal/domain"
	"goodrain.com/cloud-adaptor/internal/model"
	"goodrain.com/cloud-adaptor/internal/nsqc/producer"
	"goodrain.com/cloud-adaptor/internal/operator"
	"goodrain.com/cloud-adaptor/internal/repo"
	"goodrain.com/cloud-adaptor/internal/types"
	"goodrain.com/cloud-adaptor/pkg/bcode"
	"goodrain.com/cloud-adaptor/pkg/util/constants"
	"goodrain.com/cloud-adaptor/pkg/util/md5util"
	"goodrain.com/cloud-adaptor/pkg/util/ssh"
	"goodrain.com/cloud-adaptor/pkg/util/uuidutil"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	customClusterRepo         repo.CustomClusterRepository
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
	customClusterRepo repo.CustomClusterRepository,
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
		customClusterRepo:         customClusterRepo,
	}
}

// ListKubernetesCluster list kubernetes cluster
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

// CreateKubernetesCluster create kubernetes cluster task
func (c *ClusterUsecase) CreateKubernetesCluster(eid string, req v1.CreateKubernetesReq) (*model.CreateKubernetesTask, error) {
	if c.TaskProducer == nil {
		return nil, errors.New("TaskProducer is nil")
	}
	clusterID := uuidutil.NewUUID()
	clusterStatus := v1alpha1.OfflineState
	if req.Provider == "custom" {
		if err := c.customClusterRepo.Create(&model.CustomCluster{
			Name:         req.Name,
			EIP:          strings.Join(req.EIP, ","),
			KubeConfig:   req.KubeConfig,
			EnterpriseID: eid,
			ClusterID:    clusterID,
		}); err != nil {
			return nil, errors.Wrap(err, "create custom cluster")
		}
		kc := v1alpha1.KubeConfig{Config: req.KubeConfig}
		client, _, err := kc.GetKubeClient()
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()
			_, err := client.RESTClient().Get().AbsPath("/version").DoRaw(ctx)
			if err == nil {
				clusterStatus = v1alpha1.RunningState
			}
		}
	}

	var rkeConfig v3.RancherKubernetesEngineConfig
	if req.Provider == "rke" {
		rkeCluster := &model.RKECluster{
			Name:         req.Name,
			Stats:        v1alpha1.InitState,
			EnterpriseID: eid,
			ClusterID:    clusterID,
		}
		// Only the request to successfully create the rke cluster can send the task
		if err := c.rkeClusterRepo.Create(rkeCluster); err != nil {
			return nil, err
		}

		decRKEConfig, err := base64.StdEncoding.DecodeString(req.EncodedRKEConfig)
		if err != nil {
			return nil, errors.Wrap(bcode.ErrIncorrectRKEConfig, "decode encoded rke config")
		}
		if err := yaml.Unmarshal(decRKEConfig, &rkeConfig); err != nil {
			return nil, errors.Wrap(bcode.ErrIncorrectRKEConfig, "unmarshal rke config")
		}
		// validate nodes
		nodeList, err := c.rkeConfigToNodeList(&rkeConfig)
		if err != nil {
			return nil, err
		}
		if err := nodeList.Validate(); err != nil {
			return nil, err
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
		ClusterID:          clusterID,
	}
	if err := c.CreateKubernetesTaskRepo.Create(newTask); err != nil {
		return nil, errors.Wrap(err, "create kubernetes task")
	}
	// send task
	taskReq := types.KubernetesConfigMessage{
		EnterpriseID: eid,
		TaskID:       newTask.TaskID,
		KubernetesConfig: &v1alpha1.KubernetesClusterConfig{
			ClusterName:        newTask.Name,
			WorkerResourceType: newTask.WorkerResourceType,
			WorkerNodeNum:      newTask.WorkerNum,
			Provider:           newTask.Provider,
			Region:             newTask.Region,
			RKEConfig:          &rkeConfig,
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
	newTask.Status = clusterStatus
	return newTask, nil
}

func (c *ClusterUsecase) isAlreadyInstalled(ctx context.Context, eid, clusterID, providerName string) error {
	kubeConfig, err := c.GetKubeConfig(eid, clusterID, providerName)
	if err != nil {
		if err.Error() == "not found kube config" {
			return nil
		}
		return err
	}
	if kubeConfig == "" {
		return nil
	}

	kc := v1alpha1.KubeConfig{Config: kubeConfig}
	kubeClient, _, err := kc.GetKubeClient()
	if err != nil {
		logrus.Errorf("get kube client: %v", err)
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := kubeClient.AppsV1().Deployments(constants.Namespace).Get(ctx, "rainbond-operator", metav1.GetOptions{}); err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil
		}
		logrus.Warningf("get operator failure %s", err.Error())
		return nil
	}

	return errors.WithStack(bcode.ErrRainbondClusterInstalled)
}

func (c *ClusterUsecase) rkeConfigToNodeList(rkeConfig *v3.RancherKubernetesEngineConfig) (v1alpha1.NodeList, error) {
	if rkeConfig == nil {
		return nil, nil
	}

	var nodeList v1alpha1.NodeList
	for _, node := range rkeConfig.Nodes {
		port, err := strconv.Atoi(node.Port)
		if err != nil {
			return nil, errors.Wrapf(bcode.ErrIncorrectRKEConfig, "invalid node port %s", node.Port)
		}
		nodeList = append(nodeList, v1alpha1.ConfigNode{
			IP:               node.Address,
			InternalAddress:  node.InternalAddress,
			SSHUser:          node.User,
			SSHPort:          port,
			DockerSocketPath: node.DockerSocket,
			Roles:            node.Role,
		})
	}
	return nodeList, nil
}

// InitRainbondRegion init rainbond region
func (c *ClusterUsecase) InitRainbondRegion(ctx context.Context, eid string, req v1.InitRainbondRegionReq) (*model.InitRainbondTask, error) {
	oldTask, err := c.InitRainbondTaskRepo.GetTaskByClusterID(eid, req.Provider, req.ClusterID)
	if err != nil && !errors.Is(err, bcode.ErrInitRainbondTaskNotFound) {
		return nil, err
	}
	if oldTask != nil && !req.Retry {
		return oldTask, bcode.ErrorLastTaskNotComplete
	}

	if err := c.isAlreadyInstalled(ctx, eid, req.ClusterID, req.Provider); err != nil {
		return nil, err
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

// UpdateKubernetesCluster -
func (c *ClusterUsecase) UpdateKubernetesCluster(eid string, req v1.UpdateKubernetesReq) (*v1.UpdateKubernetesTask, error) {
	if c.TaskProducer == nil {
		logrus.Errorf("TaskProducer is nil")
		return nil, bcode.ServerErr
	}
	if req.Provider != "rke" {
		return nil, bcode.ErrNotSupportUpdateKubernetes
	}

	decodedRkeConfig, err := base64.StdEncoding.DecodeString(req.EncodedRKEConfig)
	if err != nil {
		logrus.Errorf("decode encoded rke config: %v", err)
		return nil, errors.Wrap(bcode.ErrIncorrectRKEConfig, "decode encoded rke config")
	}
	var rkeConfig v3.RancherKubernetesEngineConfig
	if err := yaml.Unmarshal(decodedRkeConfig, &rkeConfig); err != nil {
		logrus.Errorf("unmarshal rke config: %v", err)
		return nil, errors.Wrap(bcode.ErrIncorrectRKEConfig, "unmarshal rke config")
	}

	// check if the last task is complete
	version, err := c.isLastTaskComplete(eid, req.ClusterID)
	if err != nil {
		return nil, err
	}

	newTask := &model.UpdateKubernetesTask{
		TaskID:       uuidutil.NewUUID(),
		Provider:     req.Provider,
		EnterpriseID: eid,
		ClusterID:    req.ClusterID,
		NodeNumber:   len(rkeConfig.Nodes),
		Version:      version + 1, // optimistic lock
	}
	if err := c.UpdateKubernetesTaskRepo.Create(newTask); err != nil {
		return nil, errors.Wrap(err, "save update kubernetes task failure")
	}

	// send task
	taskReq := types.UpdateKubernetesConfigMessage{
		EnterpriseID: eid,
		TaskID:       newTask.TaskID,
		Config: &v1alpha1.ExpansionNode{
			Provider:     req.Provider,
			ClusterID:    req.ClusterID,
			EnterpriseID: eid,
			RKEConfig:    &rkeConfig,
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

func (c *ClusterUsecase) isLastTaskComplete(eid, clusterID string) (int, error) {
	// check if update task complete
	updateTask, err := c.UpdateKubernetesTaskRepo.GetTaskByClusterID(eid, clusterID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	if updateTask != nil && updateTask.Status != "complete" {
		return 0, errors.WithStack(bcode.ErrLastKubernetesTaskNotComplete)
	}

	// check if create task complete
	createTask, err := c.CreateKubernetesTaskRepo.GetLatestOneByClusterID(clusterID)
	if err != nil && !errors.Is(err, bcode.ErrLastTaskNotFound) {
		return 0, err
	}
	if createTask != nil && createTask.Status != "complete" {
		return 0, errors.WithStack(bcode.ErrLastKubernetesTaskNotComplete)
	}

	if updateTask != nil {
		return updateTask.Version, nil
	}

	return 0, nil
}

// GetInitRainbondTaskByClusterID get init rainbond task
func (c *ClusterUsecase) GetInitRainbondTaskByClusterID(eid, clusterID, providerName string) (*model.InitRainbondTask, error) {
	task, err := c.InitRainbondTaskRepo.GetTaskByClusterID(eid, providerName, clusterID)
	if err != nil {
		if errors.Is(err, bcode.ErrInitRainbondTaskNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// sync the status of events and the task
	c.ListTaskEvent(eid, task.TaskID)

	// get the real status from the cluster
	status, err := c.getTaskClusterStatus(task)
	if err != nil {
		logrus.Warningf("get task cluster status: %v", err)
	}
	if status == "installing" {
		task.Status = status
	}

	return task, nil
}

// GetUpdateKubernetesTask get update kubernetes task
func (c *ClusterUsecase) GetUpdateKubernetesTask(eid, clusterID, providerName string) (*v1.UpdateKubernetesRes, error) {
	var clusterName string
	if providerName == "rke" {
		cluster, err := c.rkeClusterRepo.GetCluster(eid, clusterID)
		if err != nil {
			return nil, err
		}
		clusterName = cluster.Name
	}

	task, err := c.getUpdateKubernetesTask(eid, clusterName, clusterID)
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

		rkeConfig, err := c.getRKEConfig(eid, cluster)
		if err != nil {
			return nil, err
		}
		rkeConfigBytes, err := yaml.Marshal(rkeConfig)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		re.RKEConfig = base64.StdEncoding.EncodeToString(rkeConfigBytes)

		nodeList, err := c.rkeConfigToNodeList(rkeConfig)
		if err != nil {
			return nil, err
		}
		re.NodeList = nodeList
	}

	return &re, nil
}

func (c *ClusterUsecase) getUpdateKubernetesTask(eid, name, clusterID string) (interface{}, error) {
	update, err := c.UpdateKubernetesTaskRepo.GetTaskByClusterID(eid, clusterID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if update != nil {
		return update, nil
	}

	// return create kubernetes task if exists.
	create, err := c.CreateKubernetesTaskRepo.GetLatestOneByName(name)
	if err != nil && !errors.Is(err, bcode.ErrLastTaskNotFound) {
		return nil, err
	}
	return create, nil
}

func (c *ClusterUsecase) getRKEConfig(eid string, cluster *model.RKECluster) (*v3.RancherKubernetesEngineConfig, error) {
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		configDir = "/tmp"
	}
	clusterYMLPath := fmt.Sprintf("%s/enterprise/%s/rke/%s/cluster.yml", configDir, cluster.EnterpriseID, cluster.Name)
	oldClusterYMLPath := fmt.Sprintf("%s/rke/%s/cluster.yml", configDir, cluster.Name)

	bytes, err := ioutil.ReadFile(clusterYMLPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "read rke config file")
		}
		bytes, err = ioutil.ReadFile(oldClusterYMLPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, errors.Wrap(bcode.ErrRKEConfigLost, err.Error())
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

// GetRKENodeList get rke kubernetes node list
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

// AddAccessKey add accesskey info to enterprise
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

// GetByProviderAndEnterprise get by eid
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

// CreateTaskEvent create task event
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
	ent.Reason = c.reasonFromMessage(ent.Message)

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

func (c *ClusterUsecase) reasonFromMessage(message string) string {
	if strings.Contains(message, fmt.Sprintf("namespace %s because it is being terminated", constants.Namespace)) {
		return "NamespaceBeingTerminated"
	}
	return ""
}

// ListTaskEvent list task event list
func (c *ClusterUsecase) ListTaskEvent(eid, taskID string) ([]*model.TaskEvent, error) {
	task, err := c.getTask(eid, taskID)
	if err != nil {
		if errors.Is(err, bcode.ErrClusterTaskNotFound) {
			return nil, nil
		}
		return nil, err
	}

	events, err := c.TaskEventRepo.ListEvent(eid, taskID)
	if err != nil {
		return nil, err
	}

	needSync := false
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
			needSync = true
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

	if needSync {
		if err := c.syncTaskEvents(task, events); err != nil {
			logrus.Errorf("sync task events: %v", err)
		}
	}

	return events, nil
}

func (c *ClusterUsecase) getTask(eid, taskID string) (*domain.ClusterTask, error) {
	var source interface{}
	var taskType domain.ClusterTaskType
	task := &domain.ClusterTask{}

	// init rainbond task
	initRainbondTask, err := c.InitRainbondTaskRepo.GetTask(eid, taskID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if initRainbondTask != nil {
		source = initRainbondTask
		taskType = domain.ClusterTaskTypeInitRainbond
	}

	// create kubernetes task
	createKubernetesTask, err := c.CreateKubernetesTaskRepo.GetTask(eid, taskID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if createKubernetesTask != nil {
		source = createKubernetesTask
		taskType = domain.ClusterTaskTypeCreateKubernetes
	}

	// update kubernetes cluster
	updateKubernetesCluster, err := c.UpdateKubernetesTaskRepo.GetTask(eid, taskID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if updateKubernetesCluster != nil {
		source = updateKubernetesCluster
		taskType = domain.ClusterTaskTypeUpdateKubernetes
	}

	if source == nil {
		return nil, bcode.ErrClusterTaskNotFound
	}
	mapper.Mapper(source, task)
	task.TaskType = taskType
	return task, nil
}

// GetLastCreateKubernetesTask get last create kubernetes task
func (c *ClusterUsecase) GetLastCreateKubernetesTask(eid, providerName string) (*model.CreateKubernetesTask, error) {
	task, err := c.CreateKubernetesTaskRepo.GetLastTask(eid, providerName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	updateTask, err := c.UpdateKubernetesTaskRepo.GetLastTask(eid, providerName)
	if err != nil {
		return task, nil
	}
	if task.CreatedAt.After(updateTask.CreatedAt) {
		return task, nil
	}
	cluster, err := c.rkeClusterRepo.GetCluster(eid, updateTask.ClusterID)
	if err != nil {
		return task, nil
	}
	return &model.CreateKubernetesTask{
		Name:         cluster.Name,
		Provider:     providerName,
		EnterpriseID: eid,
		TaskID:       updateTask.TaskID,
		Status:       updateTask.Status,
		ClusterID:    updateTask.ClusterID,
	}, nil
}

// GetCreateKubernetesTask get task
func (c *ClusterUsecase) GetCreateKubernetesTask(eid, taskID string) (*model.CreateKubernetesTask, error) {
	task, err := c.CreateKubernetesTaskRepo.GetTask(eid, taskID)
	if err != nil {
		updateTask, err := c.UpdateKubernetesTaskRepo.GetTask(eid, taskID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, bcode.NotFound
			}
			return nil, err
		}
		cluster, err := c.rkeClusterRepo.GetCluster(eid, updateTask.ClusterID)
		if err != nil {
			return task, nil
		}
		return &model.CreateKubernetesTask{
			Name:         cluster.Name,
			Provider:     updateTask.Provider,
			EnterpriseID: eid,
			TaskID:       updateTask.TaskID,
			Status:       updateTask.Status,
			ClusterID:    updateTask.ClusterID,
		}, nil
	}
	return task, err
}

// GetTaskRunningLists get runinig tasks
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

// GetKubeConfig get kube config file
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

// GetRegionConfig get region config
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

// UpdateInitRainbondTaskStatus update init rainbond task status
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

// DeleteKubernetesCluster delete provider
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

// GetCluster get cluster
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

// InstallCluster install cluster
func (c *ClusterUsecase) InstallCluster(eid, clusterID string) (*model.CreateKubernetesTask, error) {
	if c.TaskProducer == nil {
		logrus.Errorf("TaskProducer is nil")
		return nil, bcode.ServerErr
	}
	cluster, err := repo.NewRKEClusterRepo(c.DB).GetCluster(eid, clusterID)
	if err != nil {
		return nil, err
	}

	// check if the last task is complete
	if _, err := c.isLastTaskComplete(eid, clusterID); err != nil {
		return nil, err
	}

	newTask := &model.CreateKubernetesTask{
		Name:         cluster.Name,
		Provider:     "rke",
		EnterpriseID: eid,
		TaskID:       uuidutil.NewUUID(),
		ClusterID:    clusterID,
	}
	if err := c.CreateKubernetesTaskRepo.Create(newTask); err != nil {
		logrus.Errorf("create kubernetes task failure %s", err.Error())
		return nil, bcode.ServerErr
	}

	// get rke config
	rkeConfig, err := c.getRKEConfig(eid, cluster)
	if err != nil {
		return nil, err
	}
	// validate nodes
	nodeList, err := c.rkeConfigToNodeList(rkeConfig)
	if err != nil {
		return nil, err
	}
	if err := nodeList.Validate(); err != nil {
		return nil, err
	}

	// send task
	taskReq := types.KubernetesConfigMessage{
		EnterpriseID: eid,
		TaskID:       newTask.TaskID,
		KubernetesConfig: &v1alpha1.KubernetesClusterConfig{
			ClusterName:  newTask.Name,
			Provider:     newTask.Provider,
			EnterpriseID: eid,
			RKEConfig:    rkeConfig,
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

// SetRainbondClusterConfig set rainbond cluster config
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

// GetRainbondClusterConfig get rainbond cluster config
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

// UninstallRainbondRegion uninstall rainbond region
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

// PruneUpdateRKEConfig update rke config purely.
func (c *ClusterUsecase) PruneUpdateRKEConfig(req *v1.PruneUpdateRKEConfigReq) (*v1.PruneUpdateRKEConfigResp, error) {
	var rkeConfig *v3.RancherKubernetesEngineConfig
	if req.EncodedRKEConfig == "" {

		rkeConfig = v1alpha1.GetDefaultRKECreateClusterConfig(v1alpha1.KubernetesClusterConfig{
			Nodes: req.Nodes,
		}).(*v3.RancherKubernetesEngineConfig)
	} else {
		var rkeConfig2 v3.RancherKubernetesEngineConfig
		decodedRKEConfig, err := base64.StdEncoding.DecodeString(req.EncodedRKEConfig)
		if err != nil {
			return nil, errors.Wrapf(bcode.ErrIncorrectRKEConfig, "decode encoded rke config: %v", err)
		}
		if err := yaml.Unmarshal(decodedRKEConfig, &rkeConfig2); err != nil {
			return nil, errors.Wrapf(bcode.ErrIncorrectRKEConfig, "unmarshal rke config: %v", err)
		}
		if len(req.Nodes) > 0 {
			rkeConfig2.Nodes = c.nodeListToRKEConfigNodes(req.Nodes)
		}
		rkeConfig = &rkeConfig2
	}
	ignoreDockerVersion := true
	rkeConfig.IgnoreDockerVersion = &ignoreDockerVersion
	rkeConfigBytes, err := yaml.Marshal(rkeConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	encodedRKEConfig := base64.StdEncoding.EncodeToString(rkeConfigBytes)

	nodeList, err := c.rkeConfigToNodeList(rkeConfig)
	if err != nil {
		return nil, err
	}

	return &v1.PruneUpdateRKEConfigResp{
		Nodes:            nodeList,
		EncodedRKEConfig: encodedRKEConfig,
	}, nil
}

func (c *ClusterUsecase) nodeListToRKEConfigNodes(nodeList v1alpha1.NodeList) []v3.RKEConfigNode {
	var nodes []v3.RKEConfigNode
	for _, node := range nodeList {
		nodes = append(nodes, v3.RKEConfigNode{
			NodeName: "",
			Address:  node.IP,
			Port: func() string {
				if node.SSHPort != 0 {
					return fmt.Sprintf("%d", node.SSHPort)
				}
				return "22"
			}(),
			DockerSocket: node.DockerSocketPath,
			User: func() string {
				if node.SSHUser != "" {
					return node.SSHUser
				}
				return "docker"
			}(),
			SSHKeyPath:      "~/.ssh/id_rsa",
			Role:            node.Roles,
			InternalAddress: node.InternalAddress,
		})
	}
	return nodes
}

// GetInitNodeCmd -
func (c *ClusterUsecase) GetInitNodeCmd(ctx context.Context) (*v1.InitNodeCmdRes, error) {
	pub, err := ssh.GetOrMakeSSHRSA()
	if err != nil {
		return nil, errors.Wrap(err, "get or create ssh rsa")
	}

	if config.C.IsOffline {
		return &v1.InitNodeCmdRes{
			Cmd:       fmt.Sprintf(`export SSH_RSA="%s" && ./init_node_offline.sh`, pub),
			IsOffline: true,
		}, nil
	}
	return &v1.InitNodeCmdRes{
		Cmd: fmt.Sprintf(`export SSH_RSA="%s" && curl -sfL https://get.rainbond.com/init_node | bash`, strings.Replace(pub, "\n", "", -1)),
	}, nil
}

// ListRainbondComponents -
func (c *ClusterUsecase) ListRainbondComponents(ctx context.Context, eid, clusterID, providerName string) ([]*v1.RainbondComponent, error) {
	kubeConfig, err := c.GetKubeConfig(eid, clusterID, providerName)
	if err != nil {
		return nil, err
	}

	kc := v1alpha1.KubeConfig{Config: kubeConfig}
	kubeClient, runtimeClient, err := kc.GetKubeClient()
	if err != nil {
		return nil, errors.Wrap(bcode.ErrorKubeAPI, err.Error())
	}

	return c.listRainbondComponents(ctx, kubeClient, runtimeClient)
}

func (c *ClusterUsecase) listRainbondComponents(ctx context.Context, kubeClient kubernetes.Interface, runtimeClient client.Client) ([]*v1.RainbondComponent, error) {
	pods, err := c.listRainbondPods(ctx, kubeClient)
	if err != nil {
		return nil, err
	}

	components, err := c.listRbdComponent(ctx, runtimeClient)
	if err != nil {
		return nil, err
	}

	var res []*v1.RainbondComponent
	for _, name := range components {
		res = append(res, &v1.RainbondComponent{
			App:  name,
			Pods: pods[name],
		})
	}

	sort.Sort(v1.ByRainbondComponentPodPhase(res))
	return res, nil
}

func (c *ClusterUsecase) listRbdComponent(ctx context.Context, runtimeClient client.Client) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	components := &rainbondv1alpha1.RbdComponentList{}
	err := runtimeClient.List(ctx, components, &client.ListOptions{
		Namespace: constants.Namespace,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var appNames []string
	for _, cpt := range components.Items {
		appNames = append(appNames, cpt.Name)
	}
	appNames = append(appNames, "rainbond-operator")
	return appNames, nil
}

func (c *ClusterUsecase) listRainbondPods(ctx context.Context, kubeClient kubernetes.Interface) (map[string][]corev1.Pod, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// rainbond components
	podList, err := kubeClient.CoreV1().Pods(constants.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fields.SelectorFromSet(rbdutil.LabelsForRainbond(nil)).String(),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pods := make(map[string][]corev1.Pod)
	for _, pod := range podList.Items {
		pod := pod

		labels := pod.Labels
		appName := labels["name"]
		if len(appName) == 0 {
			logrus.Warningf("list rainbond components. label 'name' not found for pod(%s/%s)", pod.Namespace, pod.Name)
			continue
		}

		cptPods := pods[appName]
		pods[appName] = append(cptPods, pod)
	}

	// rainbond operator
	roPods, err := kubeClient.CoreV1().Pods(constants.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fields.SelectorFromSet(map[string]string{
			"release": "rainbond-operator",
		}).String(),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pods["rainbond-operator"] = roPods.Items

	return pods, nil
}

// ListPodEvents -
func (c *ClusterUsecase) ListPodEvents(ctx context.Context, eid, clusterID, providerName, podName string) ([]corev1.Event, error) {
	kubeConfig, err := c.GetKubeConfig(eid, clusterID, providerName)
	if err != nil {
		return nil, err
	}

	kc := v1alpha1.KubeConfig{Config: kubeConfig}
	kubeClient, _, err := kc.GetKubeClient()
	if err != nil {
		return nil, errors.Wrap(bcode.ErrorKubeAPI, err.Error())
	}

	return c.listPodEvents(ctx, kubeClient, podName)
}

func (c *ClusterUsecase) listPodEvents(ctx context.Context, kubeClient kubernetes.Interface, podName string) ([]corev1.Event, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	eventList, err := kubeClient.CoreV1().Events(constants.Namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", podName),
	})
	if err != nil {
		return nil, err
	}
	return eventList.Items, nil
}

func (c *ClusterUsecase) syncTaskEvents(task *domain.ClusterTask, events []*model.TaskEvent) error {
	if task.TaskType != domain.ClusterTaskTypeInitRainbond || (task.ProviderName != "rke" && task.ProviderName != "custom") {
		return nil
	}

	kubeConfig, err := c.GetKubeConfig(task.EnterpriseID, task.ClusterID, task.ProviderName)
	if err != nil {
		return err
	}

	rri := operator.NewRainbondRegionInit(v1alpha1.KubeConfig{Config: kubeConfig}, c.RainbondClusterConfigRepo)
	status, err := rri.GetRainbondRegionStatus(task.ClusterID)
	if err != nil {
		return err
	}

	var updates []string
	// update InitRainbondRegionOperator event
	if status.OperatorReady {
		event := c.getEvent("InitRainbondRegionOperator", events)
		if event != nil {
			updates = append(updates, event.EventID)
		}
	}
	// update InitRainbondRegionImageHub event
	if idx, condition := status.RainbondCluster.Status.GetCondition(rainbondv1alpha1.RainbondClusterConditionTypeImageRepository); idx != -1 && condition.Status == corev1.ConditionTrue {
		event := c.getEvent("InitRainbondRegionImageHub", events)
		if event != nil {
			updates = append(updates, event.EventID)
		}
	}
	// update InitRainbondRegionPackage event
	for _, con := range status.RainbondPackage.Status.Conditions {
		if con.Type == rainbondv1alpha1.Ready && con.Status == rainbondv1alpha1.Completed {
			event := c.getEvent("InitRainbondRegionPackage", events)
			if event != nil {
				updates = append(updates, event.EventID)
			}
		}
	}
	// update InitRainbondRegion event
	idx, condition := status.RainbondCluster.Status.GetCondition(rainbondv1alpha1.RainbondClusterConditionTypeRunning)
	if idx != -1 && condition.Status == corev1.ConditionTrue {
		event := c.getEvent("InitRainbondRegion", events)
		if event != nil {
			updates = append(updates, event.EventID)
		}
	}

	return c.TaskEventRepo.UpdateStatusInBatch(updates, "success")
}

func (c *ClusterUsecase) getEvent(stepType string, events []*model.TaskEvent) *model.TaskEvent {
	for _, event := range events {
		if event.StepType == stepType {
			return event
		}
	}
	return nil
}

func (c *ClusterUsecase) getTaskClusterStatus(task *model.InitRainbondTask) (string, error) {
	if task.Provider != "rke" && task.Provider != "custom" {
		return "", nil
	}

	kubeConfig, err := c.GetKubeConfig(task.EnterpriseID, task.ClusterID, task.Provider)
	if err != nil {
		return "", err
	}

	rri := operator.NewRainbondRegionInit(v1alpha1.KubeConfig{Config: kubeConfig}, c.RainbondClusterConfigRepo)
	status, err := rri.GetRainbondRegionStatus(task.ClusterID)
	if err != nil {
		return "", err
	}

	// update InitRainbondRegion event
	idx, condition := status.RainbondCluster.Status.GetCondition(rainbondv1alpha1.RainbondClusterConditionTypeRunning)
	if idx != -1 && condition.Status == corev1.ConditionTrue {
		return "complete", nil
	}

	if status.OperatorInstalled {
		return "installing", nil
	}

	return "", nil
}
