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

package producer

import (
	"goodrain.com/cloud-adaptor/internal/types"
	"goodrain.com/cloud-adaptor/pkg/util/constants"
)

//TaskProducer task producer
type taskChannelProducer struct {
	createQueue chan types.KubernetesConfigMessage
	initQueue   chan types.InitRainbondConfigMessage
	updateQueue chan types.UpdateKubernetesConfigMessage
}

//NewTaskChannelProducer new task channel producer
func NewTaskChannelProducer(createQueue chan types.KubernetesConfigMessage,
	initQueue chan types.InitRainbondConfigMessage,
	updateQueue chan types.UpdateKubernetesConfigMessage) TaskProducer {
	return &taskChannelProducer{
		createQueue: createQueue,
		initQueue:   initQueue,
		updateQueue: updateQueue,
	}
}

//Start start
func (c *taskChannelProducer) Start() error {
	return nil
}

//SendTask send task
func (c *taskChannelProducer) sendTask(topicName string, taskConfig interface{}) error {
	if topicName == constants.CloudCreate {
		c.createQueue <- taskConfig.(types.KubernetesConfigMessage)
	}
	if topicName == constants.CloudInit {
		c.initQueue <- taskConfig.(types.InitRainbondConfigMessage)
	}
	if topicName == constants.CloudUpdate {
		c.updateQueue <- taskConfig.(types.UpdateKubernetesConfigMessage)
	}
	return nil
}

//SendCreateKuerbetesTask send create kubernetes task
func (c *taskChannelProducer) SendCreateKuerbetesTask(config types.KubernetesConfigMessage) error {
	return c.sendTask(constants.CloudCreate, config)
}

//SendInitRainbondRegionTask send init rainbond region task
func (c *taskChannelProducer) SendInitRainbondRegionTask(config types.InitRainbondConfigMessage) error {
	return c.sendTask(constants.CloudInit, config)
}
func (c *taskChannelProducer) SendUpdateKuerbetesTask(config types.UpdateKubernetesConfigMessage) error {
	return c.sendTask(constants.CloudUpdate, config)
}

//Stop stop
func (c *taskChannelProducer) Stop() {

}
