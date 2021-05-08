package task

import (
	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
	v1 "goodrain.com/cloud-adaptor/api/cloud-adaptor/v1"
	"goodrain.com/cloud-adaptor/internal/usecase"
)

//CallBackEvent callback event
type CallBackEvent struct {
	eventProducer  *nsq.Producer
	TopicName      string
	ClusterUsecase *usecase.ClusterUsecase
}

//Event send event
func (c *CallBackEvent) Event(e v1.EventMessage) error {
	if err := c.eventProducer.Publish(c.TopicName, e.Body()); err != nil {
		if err := c.eventProducer.Publish(c.TopicName, e.Body()); err != nil {
			return err
		}
	}
	logrus.Infof("send a task %s event %+v", e.TaskID, e.Message)
	return nil
}

// HandleEvent -
func (c *CallBackEvent) HandleEvent(msg v1.EventMessage) error {
	if _, err := c.ClusterUsecase.CreateTaskEvent(&msg); err != nil {
		logrus.Errorf("save event message failure %s", err.Error())
		// return err, retry
		if err.Error() != "message is nil" {
			return err
		}
	}
	return nil
}
