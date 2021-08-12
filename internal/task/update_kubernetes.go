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
	"goodrain.com/cloud-adaptor/internal/types"
	"goodrain.com/cloud-adaptor/internal/usecase"
	"goodrain.com/cloud-adaptor/pkg/util/constants"
)

//UpdateKubernetesCluster update cluster
type UpdateKubernetesCluster struct {
	config *v1alpha1.ExpansionNode
	result chan v1.Message
}

func (c *UpdateKubernetesCluster) rollback(step, message, status string) {
	if status == "failure" {
		logrus.Errorf("%s failure, Message: %s", step, message)
	}
	c.result <- v1.Message{StepType: step, Message: message, Status: status}
}

//Run run
func (c *UpdateKubernetesCluster) Run(ctx context.Context) {
	defer c.rollback("Close", "", "")
	c.rollback("Init", "", "start")
	// create adaptor
	adaptor, err := factory.GetCloudFactory().GetRainbondClusterAdaptor(c.config.Provider, c.config.AccessKey, c.config.SecretKey)
	if err != nil {
		c.rollback("Init", fmt.Sprintf("create cloud adaptor failure %s", err.Error()), "failure")
		return
	}
	c.rollback("Init", "cloud adaptor create success", "success")
	// update cluster
	adaptor.ExpansionNode(ctx, c.config.EnterpriseID, c.config, c.rollback)
}

//GetChan get message chan
func (c *UpdateKubernetesCluster) GetChan() chan v1.Message {
	return c.result
}

type cloudUpdateTaskHandler struct {
	eventHandler *CallBackEvent
	handledTask  map[string]string
}

// NewCloudUpdateTaskHandler -
func NewCloudUpdateTaskHandler(clusterUsecase *usecase.ClusterUsecase) UpdateKubernetesTaskHandler {
	return &cloudUpdateTaskHandler{
		eventHandler: &CallBackEvent{TopicName: constants.CloudInit, ClusterUsecase: clusterUsecase},
		handledTask:  make(map[string]string),
	}
}

// HandleMsg -
func (h *cloudUpdateTaskHandler) HandleMsg(ctx context.Context, config types.UpdateKubernetesConfigMessage) error {
	if _, exist := h.handledTask[config.TaskID]; exist {
		logrus.Infof("task %s is running or complete,ignore", config.TaskID)
		return nil
	}
	initTask, err := CreateTask(UpdateKubernetesTask, config.Config, nil)
	if err != nil {
		logrus.Errorf("create task failure %s", err.Error())
		h.eventHandler.HandleEvent(config.GetEvent(&v1.Message{
			StepType: "CreateTask",
			Message:  err.Error(),
			Status:   "failure",
		}))
		return nil
	}
	// Asynchronous execution to prevent message consumption from taking too long.
	// Idempotent consumption of messages is not currently supported
	go h.run(ctx, initTask, config)
	h.handledTask[config.TaskID] = "running"
	return nil
}

// HandleMessage implements the Handler interface.
// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
func (h *cloudUpdateTaskHandler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		return nil
	}
	var initConfig types.UpdateKubernetesConfigMessage
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

func (h *cloudUpdateTaskHandler) run(ctx context.Context, initTask Task, initConfig types.UpdateKubernetesConfigMessage) {
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
	logrus.Infof("update kubernetes task %s handle success", initConfig.TaskID)
}
