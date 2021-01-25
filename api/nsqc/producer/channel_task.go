package producer

import (
	"goodrain.com/cloud-adaptor/task"
	"goodrain.com/cloud-adaptor/util/constants"
)

//TaskProducer task producer
type taskChannelProducer struct {
	createQueue chan task.KubernetesConfigMessage
	initQueue   chan task.InitRainbondConfigMessage
}

//NewTaskChannelProducer new task channel producer
func NewTaskChannelProducer(createQueue chan task.KubernetesConfigMessage, initQueue chan task.InitRainbondConfigMessage) TaskProducer {
	return &taskChannelProducer{
		createQueue: createQueue,
		initQueue:   initQueue,
	}
}

//Start start
func (c *taskChannelProducer) Start() error {
	return nil
}

//SendTask send task
func (c *taskChannelProducer) sendTask(topicName string, taskConfig interface{}) error {
	if topicName == constants.CloudCreate {
		c.createQueue <- taskConfig.(task.KubernetesConfigMessage)
	}
	if topicName == constants.CloudInit {
		c.initQueue <- taskConfig.(task.InitRainbondConfigMessage)
	}
	return nil
}

//SendCreateKuerbetesTask send create kubernetes task
func (c *taskChannelProducer) SendCreateKuerbetesTask(config task.KubernetesConfigMessage) error {
	return c.sendTask(constants.CloudCreate, config)
}

//SendInitRainbondRegionTask send init rainbond region task
func (c *taskChannelProducer) SendInitRainbondRegionTask(config task.InitRainbondConfigMessage) error {
	return c.sendTask(constants.CloudInit, config)
}

//Stop stop
func (c *taskChannelProducer) Stop() {

}
