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
