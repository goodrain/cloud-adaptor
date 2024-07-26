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

package task

import (
	"context"
	"encoding/json"
	"fmt"
	"goodrain.com/cloud-adaptor/internal/adaptor/rke"
	"k8s.io/apimachinery/pkg/fields"
	ktype "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"runtime/debug"
	"strings"
	"time"

	rainbondv1alpha1 "github.com/goodrain/rainbond-operator/api/v1alpha1"
	"github.com/nsqio/go-nsq"
	"github.com/rancher/rke/k8s"
	"github.com/sirupsen/logrus"
	apiv1 "goodrain.com/cloud-adaptor/api/cloud-adaptor/v1"
	ccv1 "goodrain.com/cloud-adaptor/api/cloud-adaptor/v1"
	"goodrain.com/cloud-adaptor/internal/adaptor/factory"
	"goodrain.com/cloud-adaptor/internal/types"
	"goodrain.com/cloud-adaptor/internal/usecase"
	"goodrain.com/cloud-adaptor/pkg/util/constants"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InitRainbondCluster init rainbond cluster
type InitRainbondCluster struct {
	config *types.InitRainbondConfig
	result chan apiv1.Message
}

func (c *InitRainbondCluster) rollback(step, message, status string) {
	if status == "failure" {
		logrus.Errorf("%s failure, Message: %s", step, message)
	}
	c.result <- apiv1.Message{StepType: step, Message: message, Status: status}
}

// CheckKubernetesStatus Check kubernetes status
func (c *InitRainbondCluster) CheckKubernetesStatus(clientset *kubernetes.Clientset) (bool, error) {
	nodeList, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	if len(nodeList.Items) == 0 {
		return false, nil
	}
	return true, err
}

func (c *InitRainbondCluster) CheckOperatorStatus(ctx context.Context, clientset *kubernetes.Clientset) error {
	//通过一个定时器来控制检测时间
	ticker := time.NewTicker(time.Second * 5)
	timer := time.NewTimer(time.Minute * 60)
	defer ticker.Stop()
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancel")
		case <-ticker.C:
		case <-timer.C:
			return nil
		}

		roPods, err := clientset.CoreV1().Pods(constants.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: fields.SelectorFromSet(map[string]string{
				"release": "rainbond",
			}).String(),
		})
		if err != nil {
			return fmt.Errorf("get rainbond-operator pod failed:%s", err)
		}

		if len(roPods.Items) == 0 {
			continue
		}
		if roPods.Items[0].Status.Phase == "Running" {
			break
		}
	}
	c.rollback("InitRainbondOperator", "", "success")
	return nil
}

func (c *InitRainbondCluster) CheckClusterStatus(ctx context.Context) error {
	adapter, _ := rke.Create()
	kubeConfig, err := adapter.GetKubeConfig(c.config.EnterpriseID, c.config.ClusterID)
	_, runtimeClient, err := kubeConfig.GetKubeClient()
	if err != nil {
		logrus.Infof("get kubeclient failure %s", err.Error())
		return err
	}

	var cluster rainbondv1alpha1.RainbondCluster
	ticker := time.NewTicker(time.Second * 10)
	timer := time.NewTimer(time.Minute * 60)
	defer timer.Stop()
	defer ticker.Stop()
	var initRainbondCluster bool
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancel")
		case <-ticker.C:
		case <-timer.C:
			return fmt.Errorf("check cluster status failure")
		}
		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel2()
		err = runtimeClient.Get(ctx2, ktype.NamespacedName{Name: constants.RainbondCluster, Namespace: constants.Namespace}, &cluster)
		if err != nil {
			logrus.Errorf("get cluster failure %s", err.Error())
			return err
		}
		//获取到cluster信息后，进行数据校验
		for _, condition := range cluster.Status.Conditions {
			if condition.Type == rainbondv1alpha1.RainbondClusterConditionTypeStorage {
				continue
			}

			status, msg := c.HandleClusterStatus(cluster, condition.Type)
			if strings.Contains(msg.Message, "not ready") {
				break
			}
			if !status && msg.Message != "" {
				c.rollback(msg.StepType, msg.Message, "failure")
				return fmt.Errorf("get clusterType %s failure%s:", msg.StepType, msg.Message)
			}
			//更新状态为成功
			c.rollback(msg.StepType, "", "success")
			logrus.Infof("get clusterType %s success", msg.StepType)
			if condition.Type == "ImageRepository" {
				initRainbondCluster = true
			}
		}
		if initRainbondCluster {
			break
		}
	}
	c.rollback("InitRainbondCluster", "", "success")

	return nil
}

func (c *InitRainbondCluster) HandleClusterStatus(cluster rainbondv1alpha1.RainbondCluster, clusterType rainbondv1alpha1.RainbondClusterConditionType) (status bool, msg ccv1.Message) {
	//如果成功就更新状态
	if idx, condition := cluster.Status.GetCondition((clusterType)); idx != -1 && condition.Status == v1.ConditionTrue {
		status = true
		msg.StepType = string(condition.Type)
	} else if condition.Status == v1.ConditionFalse {
		//	拿到这里面的一些报错信息去展示，并且退出本次安装
		msg.Status = string(condition.Status)
		msg.Message = condition.Message
		msg.StepType = string(condition.Type)
		status = false
		return
	}
	return
}

// Run run take time 214.10s
func (c *InitRainbondCluster) Run(ctx context.Context) {
	defer c.rollback("Close", "", "")
	c.rollback("Init", "", "start")
	adaptor, err := factory.GetCloudFactory().GetRainbondClusterAdaptor(c.config.Provider, c.config.AccessKey, c.config.SecretKey)
	kubeConfig, err := adaptor.GetKubeConfig(c.config.EnterpriseID, c.config.ClusterID)
	if err != nil {
		kubeConfig, err = adaptor.GetKubeConfig(c.config.EnterpriseID, c.config.ClusterID)
		if err != nil {
			logrus.Errorf("get kubeconfig failure：%s", err.Error())
			c.rollback("CheckCluster", fmt.Sprintf("get kube config failure %s", err.Error()), "failure")
			return
		}
	}
	coreClient, _, err := kubeConfig.GetKubeClient()
	if err != nil {
		c.rollback("CheckCluster", fmt.Sprintf("get kube config failure %s", err.Error()), "failure")
		return
	}

	// 检测k8s状态
	status, err := c.CheckKubernetesStatus(coreClient)
	if !status {
		c.rollback("CheckKubernetes", fmt.Sprintf("Kubernetes connection failed %s", err.Error()), "failure")
		logrus.Errorf("Kubernetes connection failed")
		return
	}
	c.rollback("CheckKubernetes", c.config.ClusterID, "success")

	//安装后检测operator的状态
	err = c.CheckOperatorStatus(ctx, coreClient)
	if err != nil {
		c.rollback("CheckOperator", fmt.Sprintf("operator check failed %s", err.Error()), "failure")
		logrus.Errorf("operator detection failed %s", err.Error())
		return
	}
	//检测cluster的状态
	err = c.CheckClusterStatus(ctx)
	if err != nil {
		logrus.Errorf("detection failed cluster: %s", err)
		return
	}

}

// GetRainbondGatewayNodeAndChaosNodes get gateway nodes
func (c *InitRainbondCluster) GetRainbondGatewayNodeAndChaosNodes(nodes []v1.Node) (gatewayNodes, chaosNodes []*rainbondv1alpha1.K8sNode) {
	for _, node := range nodes {
		if node.Annotations["rainbond.io/gateway-node"] == "true" {
			gatewayNodes = append(gatewayNodes, getK8sNode(node))
		}
		if node.Annotations["rainbond.io/chaos-node"] == "true" {
			chaosNodes = append(chaosNodes, getK8sNode(node))
		}
	}
	if len(gatewayNodes) == 0 {
		if len(nodes) < 2 {
			gatewayNodes = []*rainbondv1alpha1.K8sNode{
				getK8sNode(nodes[0]),
			}
		} else {
			gatewayNodes = []*rainbondv1alpha1.K8sNode{
				getK8sNode(nodes[0]),
				getK8sNode(nodes[1]),
			}
		}
	}
	if len(chaosNodes) == 0 {
		if len(nodes) < 2 {
			chaosNodes = []*rainbondv1alpha1.K8sNode{
				getK8sNode(nodes[0]),
			}
		} else {
			chaosNodes = []*rainbondv1alpha1.K8sNode{
				getK8sNode(nodes[0]),
				getK8sNode(nodes[1]),
			}
		}
	}
	return
}

// Stop init
func (c *InitRainbondCluster) Stop() error {
	return nil
}

// GetChan get message chan
func (c *InitRainbondCluster) GetChan() chan apiv1.Message {
	return c.result
}

func getK8sNode(node v1.Node) *rainbondv1alpha1.K8sNode {
	var Knode rainbondv1alpha1.K8sNode
	for _, address := range node.Status.Addresses {
		if address.Type == v1.NodeInternalIP {
			Knode.InternalIP = address.Address
		}
		if address.Type == v1.NodeExternalIP {
			Knode.ExternalIP = address.Address
		}
		if address.Type == v1.NodeHostName {
			Knode.Name = address.Address
		}
	}
	if externamAddress, exist := node.Annotations[k8s.ExternalAddressAnnotation]; exist && externamAddress != "" {
		logrus.Infof("set node %s externalIP %s by %s", node.Name, externamAddress, k8s.ExternalAddressAnnotation)
		Knode.ExternalIP = externamAddress
	}
	return &Knode
}

// cloudInitTaskHandler cloud init task handler
type cloudInitTaskHandler struct {
	eventHandler *CallBackEvent
	handledTask  map[string]string
}

// NewCloudInitTaskHandler -
func NewCloudInitTaskHandler(clusterUsecase *usecase.ClusterUsecase) CloudInitTaskHandler {
	return &cloudInitTaskHandler{
		eventHandler: &CallBackEvent{TopicName: constants.CloudInit, ClusterUsecase: clusterUsecase},
		handledTask:  make(map[string]string),
	}
}

// HandleMsg -
func (h *cloudInitTaskHandler) HandleMsg(ctx context.Context, initConfig types.InitRainbondConfigMessage) error {
	if _, exist := h.handledTask[initConfig.TaskID]; exist {
		logrus.Infof("task %s is running or complete,ignore", initConfig.TaskID)
		return nil
	}
	initTask, err := CreateTask(InitRainbondClusterTask, initConfig.InitRainbondConfig)
	if err != nil {
		logrus.Errorf("create task failure %s", err.Error())
		h.eventHandler.HandleEvent(initConfig.GetEvent(&apiv1.Message{
			StepType: "CreateTask",
			Message:  err.Error(),
			Status:   "failure",
		}))
		return nil
	}
	// Asynchronous execution to prevent message consumption from taking too long.
	// Idempotent consumption of messages is not currently supported
	go h.run(ctx, initTask, initConfig)
	h.handledTask[initConfig.TaskID] = "running"
	return nil
}

// HandleMessage implements the Handler interface.
// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
func (h *cloudInitTaskHandler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		return nil
	}
	var initConfig types.InitRainbondConfigMessage
	if err := json.Unmarshal(m.Body, &initConfig); err != nil {
		logrus.Errorf("unmarshal init rainbond config message failure %s", err.Error())
		return nil
	}
	if err := h.HandleMsg(context.Background(), initConfig); err != nil {
		logrus.Errorf("handle init rainbond config message failure %s", err.Error())
		return nil
	}
	return nil
}

func (h *cloudInitTaskHandler) run(ctx context.Context, initTask Task, initConfig types.InitRainbondConfigMessage) {
	defer func() {
		h.handledTask[initConfig.TaskID] = "complete"
	}()
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()
	closeChan := make(chan struct{})
	go func() {
		defer close(closeChan)
		for message := range initTask.GetChan() {
			if message.StepType == "Close" {
				return
			}
			h.eventHandler.HandleEvent(initConfig.GetEvent(&message))
		}
	}()
	initTask.Run(ctx)
	//waiting message handle complete
	<-closeChan
	logrus.Infof("init rainbond region task %s handle success", initConfig.TaskID)
}
