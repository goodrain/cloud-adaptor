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
