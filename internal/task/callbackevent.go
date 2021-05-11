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
