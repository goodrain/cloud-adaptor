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
	"context"

	"goodrain.com/cloud-adaptor/internal/task"
	"goodrain.com/cloud-adaptor/internal/types"
)

// Consumer -
type taskChannelConsumer struct {
	ctx                         context.Context
	createQueue                 chan types.KubernetesConfigMessage
	initQueue                   chan types.InitRainbondConfigMessage
	updateQueue                 chan types.UpdateKubernetesConfigMessage
	createKubernetesTaskHandler task.CreateKubernetesTaskHandler
	cloudInitTaskHandler        task.CloudInitTaskHandler
	cloudUpdateTaskHandler      task.UpdateKubernetesTaskHandler
}

// NewTaskChannelConsumer creates a new consumer.
func NewTaskChannelConsumer(
	ctx context.Context,
	createQueue chan types.KubernetesConfigMessage,
	initQueue chan types.InitRainbondConfigMessage,
	updateQueue chan types.UpdateKubernetesConfigMessage,
	createHandler task.CreateKubernetesTaskHandler,
	initHandler task.CloudInitTaskHandler,
	cloudUpdateTaskHandler task.UpdateKubernetesTaskHandler,
) TaskConsumer {
	return &taskChannelConsumer{
		ctx:                         ctx,
		createQueue:                 createQueue,
		initQueue:                   initQueue,
		updateQueue:                 updateQueue,
		createKubernetesTaskHandler: createHandler,
		cloudInitTaskHandler:        initHandler,
		cloudUpdateTaskHandler:      cloudUpdateTaskHandler,
	}
}

// Start -
func (c *taskChannelConsumer) Start() error {
	for {
		select {
		case <-c.ctx.Done():
			return nil
		case createMsg := <-c.createQueue:
			c.createKubernetesTaskHandler.HandleMsg(c.ctx, createMsg)
		case initMsg := <-c.initQueue:
			c.cloudInitTaskHandler.HandleMsg(c.ctx, initMsg)
		case updateMsg := <-c.updateQueue:
			c.cloudUpdateTaskHandler.HandleMsg(c.ctx, updateMsg)
		}
	}
}
