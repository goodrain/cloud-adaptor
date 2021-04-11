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
