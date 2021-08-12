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

package types

import (
	v1 "goodrain.com/cloud-adaptor/api/cloud-adaptor/v1"
	"goodrain.com/cloud-adaptor/internal/adaptor/v1alpha1"
)

//InitRainbondConfig init rainbond region config
type InitRainbondConfig struct {
	EnterpriseID string `json:"enterprise_id"`
	ClusterID    string `json:"cluster_id"`
	AccessKey    string `json:"access_key"`
	SecretKey    string `json:"secret_key"`
	Provider     string `json:"provider"`
	TaskID       string `json:"task_id"`
}

//KubernetesConfigMessage nsq message
type KubernetesConfigMessage struct {
	EnterpriseID     string                            `json:"enterprise_id,omitempty"`
	TaskID           string                            `json:"task_id,omitempty"`
	KubernetesConfig *v1alpha1.KubernetesClusterConfig `json:"kubernetes_config,omitempty"`
}

//UpdateKubernetesConfigMessage -
type UpdateKubernetesConfigMessage struct {
	EnterpriseID string                  `json:"enterprise_id,omitempty"`
	TaskID       string                  `json:"task_id,omitempty"`
	Config       *v1alpha1.ExpansionNode `json:"config,omitempty"`
}

//InitRainbondConfigMessage nsq message
type InitRainbondConfigMessage struct {
	EnterpriseID       string              `json:"enterprise_id,omitempty"`
	TaskID             string              `json:"task_id,omitempty"`
	InitRainbondConfig *InitRainbondConfig `json:"init_rainbond_config,omitempty"`
}

//GetEvent get event
func (i InitRainbondConfigMessage) GetEvent(m *v1.Message) v1.EventMessage {
	return v1.EventMessage{
		EnterpriseID: i.EnterpriseID,
		TaskID:       i.TaskID,
		Message:      m,
	}
}

//GetEvent get event
func (i KubernetesConfigMessage) GetEvent(m *v1.Message) v1.EventMessage {
	return v1.EventMessage{
		EnterpriseID: i.EnterpriseID,
		TaskID:       i.TaskID,
		Message:      m,
	}
}

//GetEvent get event
func (i UpdateKubernetesConfigMessage) GetEvent(m *v1.Message) v1.EventMessage {
	return v1.EventMessage{
		EnterpriseID: i.EnterpriseID,
		TaskID:       i.TaskID,
		Message:      m,
	}
}
