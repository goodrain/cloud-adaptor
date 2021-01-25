// RAINBOND, Application Management Platform
// Copyright (C) 2020-2020 Goodrain Co., Ltd.

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

package cmds

import (
	"testing"
	"time"

	nsq "github.com/nsqio/go-nsq"
	"goodrain.com/cloud-adaptor/task"
)

func TestSendEvent(t *testing.T) {
	config := nsq.NewConfig()
	eventProducer, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		t.Fatal(err)
	}
	defer eventProducer.Stop()
	eventHandler := &CallBackEvent{
		eventProducer: eventProducer,
		TopicName:     "cloud-event",
	}
	// eventHandler.Event(task.EventMessage{
	// 	EnterpriseID: "73376e85bd09060430d4bb61ba9f6612",
	// 	TaskID:       "a4c935a242c1445ba4db8ca643f3d629",
	// 	Message: &task.Message{
	// 		StepType: "Init",
	// 		Message:  "test",
	// 		Status:   "success",
	// 	},
	// })
	// time.Sleep(time.Second * 2)
	// eventHandler.Event(task.EventMessage{
	// 	EnterpriseID: "73376e85bd09060430d4bb61ba9f6612",
	// 	TaskID:       "a4c935a242c1445ba4db8ca643f3d629",
	// 	Message: &task.Message{
	// 		StepType: "CreateVSWitch",
	// 		Message:  "test",
	// 		Status:   "success",
	// 	},
	// })
	// time.Sleep(time.Second * 4)
	// eventHandler.Event(task.EventMessage{
	// 	EnterpriseID: "73376e85bd09060430d4bb61ba9f6612",
	// 	TaskID:       "a4c935a242c1445ba4db8ca643f3d629",
	// 	Message: &task.Message{
	// 		StepType: "AllocateResource",
	// 		Message:  "test",
	// 		Status:   "start",
	// 	},
	// })
	// eventHandler.Event(task.EventMessage{
	// 	EnterpriseID: "73376e85bd09060430d4bb61ba9f6612",
	// 	TaskID:       "a4c935a242c1445ba4db8ca643f3d629",
	// 	Message: &task.Message{
	// 		StepType: "AllocateResource",
	// 		Message:  "test",
	// 		Status:   "success",
	// 	},
	// })
	// time.Sleep(time.Second * 1)
	// eventHandler.Event(task.EventMessage{
	// 	EnterpriseID: "73376e85bd09060430d4bb61ba9f6612",
	// 	TaskID:       "a4c935a242c1445ba4db8ca643f3d629",
	// 	Message: &task.Message{
	// 		StepType: "CreateCluster",
	// 		Message:  "test",
	// 		Status:   "start",
	// 	},
	// })
	// time.Sleep(time.Second * 2)
	eventHandler.Event(task.EventMessage{
		EnterpriseID: "73376e85bd09060430d4bb61ba9f6612",
		TaskID:       "f2ab56ed612c4d9594e086327d8c22b7",
		Message: &task.Message{
			StepType: "InitRainbondRegionRegionConfig",
			Message:  "",
			Status:   "success",
		},
	})
	eventHandler.Event(task.EventMessage{
		EnterpriseID: "73376e85bd09060430d4bb61ba9f6612",
		TaskID:       "f2ab56ed612c4d9594e086327d8c22b7",
		Message: &task.Message{
			StepType: "InitRainbondRegion",
			Message:  "",
			Status:   "success",
		},
	})
	time.Sleep(time.Second * 3)
}
