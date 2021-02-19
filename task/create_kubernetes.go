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

	"github.com/sirupsen/logrus"
	"goodrain.com/cloud-adaptor/adaptor/factory"
	"goodrain.com/cloud-adaptor/adaptor/v1alpha1"
	v1 "goodrain.com/cloud-adaptor/api/openapi/types/v1"
)

//CreateKubernetesCluster create cluster
type CreateKubernetesCluster struct {
	config *v1alpha1.KubernetesClusterConfig
	result chan v1.Message
}

func (c *CreateKubernetesCluster) rollback(step, message, status string) {
	if status == "failure" {
		logrus.Errorf("%s failure, Message: %s", step, message)
	}
	c.result <- v1.Message{StepType: step, Message: message, Status: status}
}

//Run run
func (c *CreateKubernetesCluster) Run(ctx context.Context) {
	defer c.rollback("Close", "", "")
	c.rollback("Init", "", "start")
	// create adaptor
	adaptor, err := factory.GetCloudFactory().GetRainbondClusterAdaptor(c.config.Provider, c.config.AccessKey, c.config.SecretKey)
	if err != nil {
		c.rollback("Init", fmt.Sprintf("create cloud adaptor failure %s", err.Error()), "failure")
		return
	}
	c.rollback("Init", "cloud adaptor create success", "success")
	// create cluster
	adaptor.CreateRainbondKubernetes(ctx, c.config, c.rollback)
}

//GetChan get message chan
func (c *CreateKubernetesCluster) GetChan() chan v1.Message {
	return c.result
}
