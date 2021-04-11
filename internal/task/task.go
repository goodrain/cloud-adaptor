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

package task

import (
	"context"
	"fmt"

	"github.com/google/wire"
	v1 "goodrain.com/cloud-adaptor/api/cloud-adaptor/v1"
	"goodrain.com/cloud-adaptor/internal/adaptor/v1alpha1"
	"goodrain.com/cloud-adaptor/internal/types"
)

// ProviderSet is task providers.
var ProviderSet = wire.NewSet(NewCreateKubernetesTaskHandler, NewCloudInitTaskHandler, NewCloudUpdateTaskHandler)

//Task Asynchronous tasks
type Task interface {
	Run(ctx context.Context)
	GetChan() chan v1.Message
}

//Type task type
type Type string

//CreateKubernetesTask create kubernetes task
var CreateKubernetesTask Type = "create_kubernetes"

//UpdateKubernetesTask update kubernetes task
var UpdateKubernetesTask Type = "update_kubernetes"

//InitRainbondClusterTask init rainbond cluster task
var InitRainbondClusterTask Type = "init_rainbond_cluster"

//CreateTask create task
func CreateTask(taskType Type, config interface{}) (Task, error) {
	switch taskType {
	case CreateKubernetesTask:
		cconfig, ok := config.(*v1alpha1.KubernetesClusterConfig)
		if !ok {
			return nil, fmt.Errorf("config must be *v1alpha1.KubernetesClusterConfig")
		}
		return &CreateKubernetesCluster{result: make(chan v1.Message, 10), config: cconfig}, nil
	case InitRainbondClusterTask:
		cconfig, ok := config.(*types.InitRainbondConfig)
		if !ok {
			return nil, fmt.Errorf("config must be *InitRainbondConfig")
		}
		return &InitRainbondCluster{result: make(chan v1.Message, 10), config: cconfig}, nil
	case UpdateKubernetesTask:
		cconfig, ok := config.(*v1alpha1.ExpansionNode)
		if !ok {
			return nil, fmt.Errorf("config must be *v1alpha1.ExpansionNode")
		}
		return &UpdateKubernetesCluster{result: make(chan v1.Message, 10), config: cconfig}, nil
	}
	return nil, fmt.Errorf("task type not support")
}
