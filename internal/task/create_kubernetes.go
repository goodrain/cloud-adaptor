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
	"runtime/debug"

	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
	v1 "goodrain.com/cloud-adaptor/api/cloud-adaptor/v1"
	"goodrain.com/cloud-adaptor/internal/adaptor/factory"
	"goodrain.com/cloud-adaptor/internal/adaptor/v1alpha1"
	"goodrain.com/cloud-adaptor/internal/usecase"
	"goodrain.com/cloud-adaptor/internal/types"
	"goodrain.com/cloud-adaptor/pkg/util/constants"
)

//CreateKubernetesCluster create cluster
type CreateKubernetesCluster struct {
	config *v1alpha1.KubernetesClusterConfig
	result chan v1.Message
}

func (c *CreateKubernetesCluster) rollback(step, message, status string) {
	if status == "failure" {
		logrus.Errorf("%s failure, Message: %s", step, message)
	}
	c.result <- v1.Message{StepType: step, Message: message, Status: status}
}

//Run run
func (c *CreateKubernetesCluster) Run(ctx context.Context) {
	defer c.rollback("Close", "", "")
	c.rollback("Init", "", "start")
	// create adaptor
	adaptor, err := factory.GetCloudFactory().GetRainbondClusterAdaptor(c.config.Provider, c.config.AccessKey, c.config.SecretKey)
	if err != nil {
		c.rollback("Init", fmt.Sprintf("create cloud adaptor failure %s", err.Error()), "failure")
		return
	}
	c.rollback("Init", "cloud adaptor create success", "success")
	// create cluster
	adaptor.CreateRainbondKubernetes(ctx, c.config.EnterpriseID, c.config, c.rollback)
}

//GetChan get message chan
func (c *CreateKubernetesCluster) GetChan() chan v1.Message {
	return c.result
}

//createKubernetesTaskHandler create kubernetes task handler
type createKubernetesTaskHandler struct {
	eventHandler *CallBackEvent
}

// NewCreateKubernetesTaskHandler -
func NewCreateKubernetesTaskHandler(clusterUsecase *usecase.ClusterUsecase) CreateKubernetesTaskHandler {
	return &createKubernetesTaskHandler{
		eventHandler: &CallBackEvent{
			TopicName:      constants.CloudCreate,
			ClusterUsecase: clusterUsecase,
		},
	}
}

// HandleMsg -
func (h *createKubernetesTaskHandler) HandleMsg(ctx context.Context, createConfig types.KubernetesConfigMessage) error {
	initTask, err := CreateTask(CreateKubernetesTask, createConfig.KubernetesConfig, nil)
	if err != nil {
		logrus.Errorf("create task failure %s", err.Error())
		h.eventHandler.HandleEvent(createConfig.GetEvent(&v1.Message{
			StepType: "CreateTask",
			Message:  err.Error(),
			Status:   "failure",
		}))
		return nil
	}
	go h.run(ctx, initTask, createConfig)
	return nil
}

// HandleMessage implements the Handler interface.
// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
func (h *createKubernetesTaskHandler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		return nil
	}
	var createConfig types.KubernetesConfigMessage
	if err := json.Unmarshal(m.Body, &createConfig); err != nil {
		logrus.Errorf("unmarshal create kubernetes config message failure %s", err.Error())
		return nil
	}
	if err := h.HandleMsg(context.Background(), createConfig); err != nil {
		logrus.Errorf("handle create kubernetes config message failure %s", err.Error())
		return nil
	}
	return nil
}

func (h *createKubernetesTaskHandler) run(ctx context.Context, initTask Task, createConfig types.KubernetesConfigMessage) {
	closeChan := make(chan struct{})
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()
	go func() {
		defer close(closeChan)
		for message := range initTask.GetChan() {
			if message.StepType == "Close" {
				return
			}
			h.eventHandler.HandleEvent(createConfig.GetEvent(&message))
		}
	}()
	initTask.Run(ctx)
	//waiting message handle complete
	<-closeChan
	logrus.Infof("create kubernetes task %s handle success", createConfig.TaskID)
}
