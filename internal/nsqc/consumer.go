// RAINBOND, Application Management Platform
// Copyright (C) 2020-2021 Goodrain Co., Ltd.

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

package nsqc

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/nsqio/go-nsq"
	"github.com/prometheus/common/log"
	"goodrain.com/cloud-adaptor/cmd/cloud-adaptor/config"
	"goodrain.com/cloud-adaptor/internal/task"
	"goodrain.com/cloud-adaptor/pkg/util/constants"
)

//TaskConsumer task producer
type TaskConsumer interface {
	Start() error
}

// taskConsumer -
type taskConsumer struct {
	config                      *config.Config
	createKubernetesTaskHandler task.CreateKubernetesTaskHandler
	cloudInitTaskHandler        task.CloudInitTaskHandler
}

// NewTaskConsumer creates a new consumer.
func NewTaskConsumer(config *config.Config, createHandler task.CreateKubernetesTaskHandler, initHandler task.CloudInitTaskHandler) TaskConsumer {
	return &taskConsumer{
		config:                      config,
		createKubernetesTaskHandler: createHandler,
		cloudInitTaskHandler:        initHandler,
	}
}

func (c *taskConsumer) Start() error {
	config := nsq.NewConfig()
	//cloud init handler
	initConsumer, err := nsq.NewConsumer(constants.CloudInit, "default", config)
	if err != nil {
		return err
	}

	// Set the Handler for messages received by this Consumer. Can be called multiple times.
	// See also AddConcurrentHandlers.
	initConsumer.AddHandler(c.cloudInitTaskHandler)

	// Use nsqlookupd to discover nsqd instances.
	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
	err = initConsumer.ConnectToNSQLookupd(c.config.NSQConfig.NsqLookupdAddress)
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
	createConsumer.AddHandler(c.createKubernetesTaskHandler)

	// Use nsqlookupd to discover nsqd instances.
	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
	err = createConsumer.ConnectToNSQLookupd(c.config.NSQConfig.NsqLookupdAddress)
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
