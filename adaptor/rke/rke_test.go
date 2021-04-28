package rke

import (
	"context"
	"fmt"
	"testing"

	"goodrain.com/cloud-adaptor/adaptor/v1alpha1"
	"goodrain.com/cloud-adaptor/api/infrastructure/datastore"
)

func TestCreateCluster(t *testing.T) {
	rke := &rkeAdaptor{}
	rke.CreateRainbondKubernetes(context.TODO(), "test", &v1alpha1.KubernetesClusterConfig{
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

func TestExpansionNode(t *testing.T) {
	rke := &rkeAdaptor{
		Repo: NewRKEClusterRepo(datastore.NewDB()),
	}
	rke.ExpansionNode(context.TODO(), "test", &v1alpha1.ExpansionNode{
		ClusterID: "local-104",
		Nodes: v1alpha1.NodeList{
			{
				IP:    "192.168.56.104",
				Roles: []string{"controlplane", "etcd", "worker"},
			},
			{
				IP:    "192.168.56.103",
				Roles: []string{"controlplane", "worker"},
			},
		},
	}, func(step, message, status string) {
		fmt.Printf("%s\t%s\t%s\n", step, message, status)
	})
}
