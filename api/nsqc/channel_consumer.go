package nsqc

import (
	"context"

	"goodrain.com/cloud-adaptor/api/handler"
	"goodrain.com/cloud-adaptor/task"
)

// Consumer -
type taskChannelConsumer struct {
	ctx                         context.Context
	createQueue                 chan task.KubernetesConfigMessage
	initQueue                   chan task.InitRainbondConfigMessage
	createKubernetesTaskHandler handler.CreateKubernetesTaskHandler
	cloudInitTaskHandler        handler.CloudInitTaskHandler
}

// NewTaskChannelConsumer creates a new consumer.
func NewTaskChannelConsumer(ctx context.Context, createQueue chan task.KubernetesConfigMessage, initQueue chan task.InitRainbondConfigMessage, createHandler handler.CreateKubernetesTaskHandler, initHandler handler.CloudInitTaskHandler) TaskConsumer {
	return &taskChannelConsumer{
		ctx:                         ctx,
		createQueue:                 createQueue,
		initQueue:                   initQueue,
		createKubernetesTaskHandler: createHandler,
		cloudInitTaskHandler:        initHandler,
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
		}
	}
}
