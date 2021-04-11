package nsqc

import (
	"github.com/google/wire"
	"goodrain.com/cloud-adaptor/internal/nsqc/producer"
)

// ProviderSet is mq providers.
var ProviderSet = wire.NewSet(producer.NewTaskChannelProducer)
