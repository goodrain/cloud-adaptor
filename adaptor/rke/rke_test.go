package rke

import (
	"context"
	"fmt"
	"testing"

	"goodrain.com/cloud-adaptor/adaptor/v1alpha1"
)

func TestCreateCluster(t *testing.T) {
	rke := &rkeAdaptor{}
	rke.CreateRainbondKubernetes(context.TODO(), &v1alpha1.KubernetesClusterConfig{
		Nodes: []v1alpha1.ConfigNode{
			{
				IP:    "192.168.56.104",
				Roles: []string{"controlplane", "etcd", "worker"},
			},
		},
	}, func(step, message, status string) {
		fmt.Printf("%s\t%s\t%s\n", step, message, status)
	})
	// config := v1alpha1.GetDefaultRKECreateClusterConfig([]v3.RKEConfigNode{
	// 	{
	// 		Address:      "192.168.56.103",
	// 		Port:         "22",
	// 		Role:         []string{"worker", "etcd", "controlplane"},
	// 		DockerSocket: "/var/run/docker.sock",
	// 		SSHKeyPath:   "~/.ssh/id_rsa",
	// 	},
	// })
	// rke.CreateCluster(config)
}
