package task

import (
	"context"

	nsq "github.com/nsqio/go-nsq"
	"goodrain.com/cloud-adaptor/internal/types"
)

// CloudInitTaskHandler -
type CloudInitTaskHandler interface {
	HandleMsg(ctx context.Context, initConfig types.InitRainbondConfigMessage) error
	HandleMessage(m *nsq.Message) error
}

//CreateKubernetesTaskHandler create kubernetes task handler
type CreateKubernetesTaskHandler interface {
	HandleMsg(ctx context.Context, createConfig types.KubernetesConfigMessage) error
	HandleMessage(m *nsq.Message) error
}

//UpdateKubernetesTaskHandler -
type UpdateKubernetesTaskHandler interface {
	HandleMsg(ctx context.Context, createConfig types.UpdateKubernetesConfigMessage) error
	HandleMessage(m *nsq.Message) error
}
