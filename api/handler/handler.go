package handler

import (
	"context"

	nsq "github.com/nsqio/go-nsq"
	"goodrain.com/cloud-adaptor/task"
)

// CloudInitTaskHandler -
type CloudInitTaskHandler interface {
	HandleMsg(ctx context.Context, initConfig task.InitRainbondConfigMessage) error
	HandleMessage(m *nsq.Message) error
}

//CreateKubernetesTaskHandler create kubernetes task handler
type CreateKubernetesTaskHandler interface {
	HandleMsg(ctx context.Context, createConfig task.KubernetesConfigMessage) error
	HandleMessage(m *nsq.Message) error
}
