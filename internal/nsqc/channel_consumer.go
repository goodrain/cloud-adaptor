package nsqc

import (
	"context"

	"goodrain.com/cloud-adaptor/api/handler"
	"goodrain.com/cloud-adaptor/internal/task"
)

// Consumer -
type taskChannelConsumer struct {
	ctx                         context.Context
	createQueue                 chan task.KubernetesConfigMessage
	initQueue                   chan task.InitRainbondConfigMessage
	updateQueue                 chan task.UpdateKubernetesConfigMessage
	createKubernetesTaskHandler handler.CreateKubernetesTaskHandler
	cloudInitTaskHandler        handler.CloudInitTaskHandler
	cloudUpdateTaskHandler      handler.UpdateKubernetesTaskHandler
}

// NewTaskChannelConsumer creates a new consumer.
func NewTaskChannelConsumer(
	ctx context.Context,
	createQueue chan task.KubernetesConfigMessage,
	initQueue chan task.InitRainbondConfigMessage,
	updateQueue chan task.UpdateKubernetesConfigMessage,
	createHandler handler.CreateKubernetesTaskHandler,
	initHandler handler.CloudInitTaskHandler,
	cloudUpdateTaskHandler handler.UpdateKubernetesTaskHandler,
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
