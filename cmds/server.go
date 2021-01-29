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

package cmds

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	nsq "github.com/nsqio/go-nsq"
	"github.com/prometheus/common/log"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"goodrain.com/cloud-adaptor/api/cluster"
	"goodrain.com/cloud-adaptor/api/handler"
	"goodrain.com/cloud-adaptor/task"
	"goodrain.com/cloud-adaptor/util/constants"
)

var serverCommand = &cli.Command{
	Name: "server",
	Flags: []cli.Flag{&cli.StringFlag{
		Name:    "nsqd-server",
		Aliases: []string{"nsqd"},
		Value:   "127.0.0.1:4150",
		Usage:   "nsqd server address",
	}, &cli.StringFlag{
		Name:    "nsq-lookupd-server",
		Aliases: []string{"lookupd"},
		Value:   "127.0.0.1:4161",
		Usage:   "nsq lookupd server address",
	}},
	Action: runServer,
	Usage:  "run cloud adaptor server",
}

func runServer(ctx *cli.Context) error {
	config := nsq.NewConfig()
	//create event callback
	eventProducer, err := nsq.NewProducer(ctx.String("nsqd-server"), config)
	if err != nil {
		return err
	}
	defer eventProducer.Stop()
	for {
		if err := eventProducer.Ping(); err != nil {
			logrus.Errorf("ping nsqd server %s failure %s", ctx.String("nsqd-server"), err.Error())
			time.Sleep(time.Second * 2)
			continue
		}
		logrus.Infof("ping nsqd server %s success", ctx.String("nsqd-server"))
		break
	}
	eventHandler := &CallBackEvent{
		eventProducer: eventProducer,
		TopicName:     "cloud-event",
	}
	//cloud init handler
	initConsumer, err := nsq.NewConsumer(constants.CloudInit, "default", config)
	if err != nil {
		return err
	}

	// Set the Handler for messages received by this Consumer. Can be called multiple times.
	// See also AddConcurrentHandlers.
	initConsumer.AddHandler(&cloudInitTaskHandler{
		eventHandler: eventHandler,
		handledTask:  make(map[string]string),
	})

	// Use nsqlookupd to discover nsqd instances.
	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
	err = initConsumer.ConnectToNSQLookupd(ctx.String("nsq-lookupd-server"))
	if err != nil {
		return err
	}
	defer initConsumer.Stop()

	createConsumer, err := nsq.NewConsumer(constants.CloudCreate, "default", config)
	if err != nil {
		return err
	}
	// Set the Handler for messages received by this Consumer. Can be called multiple times.
	// See also AddConcurrentHandlers.
	createConsumer.AddHandler(&createKubernetesTaskHandler{
		eventHandler: eventHandler,
	})

	// Use nsqlookupd to discover nsqd instances.
	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
	err = createConsumer.ConnectToNSQLookupd(ctx.String("nsq-lookupd-server"))
	if err != nil {
		return err
	}

	// Gracefully stop the consumer.
	defer createConsumer.Stop()
	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	select {
	case v := <-initConsumer.StopChan:
		log.Errorf("Received initConsumer stop signal %d, exiting gracefully...", v)
	case v := <-createConsumer.StopChan:
		log.Errorf("Received createConsumer stop signal %d, exiting gracefully...", v)
	case <-term:
		log.Warn("Received SIGTERM, exiting gracefully...")
	}
	log.Info("See you next time!")
	return nil
}

//cloudInitTaskHandler cloud init task handler
type cloudInitTaskHandler struct {
	eventHandler *CallBackEvent
	handledTask  map[string]string
}

// NewCloudInitTaskHandler -
func NewCloudInitTaskHandler(clusterUsecase cluster.Usecase) handler.CloudInitTaskHandler {
	return &cloudInitTaskHandler{
		eventHandler: &CallBackEvent{TopicName: constants.CloudInit, ClusterUsecase: clusterUsecase},
		handledTask:  make(map[string]string),
	}
}

// HandleMsg -
func (h *cloudInitTaskHandler) HandleMsg(ctx context.Context, initConfig task.InitRainbondConfigMessage) error {
	if _, exist := h.handledTask[initConfig.TaskID]; exist {
		logrus.Infof("task %s is running or complete,ignore", initConfig.TaskID)
		return nil
	}
	initTask, err := task.CreateTask(task.InitRainbondClusterTask, initConfig.InitRainbondConfig)
	if err != nil {
		logrus.Errorf("create task failure %s", err.Error())
		h.eventHandler.HandleEvent(initConfig.GetEvent(&task.Message{
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
	var initConfig task.InitRainbondConfigMessage
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

func (h *cloudInitTaskHandler) run(ctx context.Context, initTask task.Task, initConfig task.InitRainbondConfigMessage) {
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

//createKubernetesTaskHandler create kubernetes task handler
type createKubernetesTaskHandler struct {
	eventHandler *CallBackEvent
}

// NewCreateKubernetesTaskHandler -
func NewCreateKubernetesTaskHandler(clusterUsecase cluster.Usecase) handler.CreateKubernetesTaskHandler {
	return &createKubernetesTaskHandler{
		eventHandler: &CallBackEvent{TopicName: constants.CloudCreate, ClusterUsecase: clusterUsecase},
	}
}

// HandleMsg -
func (h *createKubernetesTaskHandler) HandleMsg(ctx context.Context, createConfig task.KubernetesConfigMessage) error {
	initTask, err := task.CreateTask(task.CreateKubernetesTask, createConfig.KubernetesConfig)
	if err != nil {
		logrus.Errorf("create task failure %s", err.Error())
		h.eventHandler.HandleEvent(createConfig.GetEvent(&task.Message{
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
	var createConfig task.KubernetesConfigMessage
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

func (h *createKubernetesTaskHandler) run(ctx context.Context, initTask task.Task, createConfig task.KubernetesConfigMessage) {
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

//CallBackEvent callback event
type CallBackEvent struct {
	eventProducer  *nsq.Producer
	TopicName      string
	ClusterUsecase cluster.Usecase `inject:""`
}

//Event send event
func (c *CallBackEvent) Event(e task.EventMessage) error {
	if err := c.eventProducer.Publish(c.TopicName, e.Body()); err != nil {
		if err := c.eventProducer.Publish(c.TopicName, e.Body()); err != nil {
			return err
		}
	}
	logrus.Infof("send a task %s event %+v", e.TaskID, e.Message)
	return nil
}

// HandleEvent -
func (c *CallBackEvent) HandleEvent(msg task.EventMessage) error {
	if _, err := c.ClusterUsecase.CreateTaskEvent(&msg); err != nil {
		logrus.Errorf("save event message failure %s", err.Error())
		// return err, retry
		if err.Error() != "message is nil" {
			return err
		}
	}
	return nil
}

type cloudUpdateTaskHandler struct {
	eventHandler *CallBackEvent
	handledTask  map[string]string
}

// NewCloudUpdateTaskHandler -
func NewCloudUpdateTaskHandler(clusterUsecase cluster.Usecase) handler.UpdateKubernetesTaskHandler {
	return &cloudUpdateTaskHandler{
		eventHandler: &CallBackEvent{TopicName: constants.CloudInit, ClusterUsecase: clusterUsecase},
		handledTask:  make(map[string]string),
	}
}

// HandleMsg -
func (h *cloudUpdateTaskHandler) HandleMsg(ctx context.Context, config task.UpdateKubernetesConfigMessage) error {
	if _, exist := h.handledTask[config.TaskID]; exist {
		logrus.Infof("task %s is running or complete,ignore", config.TaskID)
		return nil
	}
	initTask, err := task.CreateTask(task.UpdateKubernetesTask, config.Config)
	if err != nil {
		logrus.Errorf("create task failure %s", err.Error())
		h.eventHandler.HandleEvent(config.GetEvent(&task.Message{
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
	var initConfig task.UpdateKubernetesConfigMessage
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

func (h *cloudUpdateTaskHandler) run(ctx context.Context, initTask task.Task, initConfig task.UpdateKubernetesConfigMessage) {
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
