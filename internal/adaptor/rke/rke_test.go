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

package rke

import (
	"context"
	"fmt"
	"goodrain.com/cloud-adaptor/internal/repo"
	"testing"

	"goodrain.com/cloud-adaptor/internal/adaptor/v1alpha1"
	"goodrain.com/cloud-adaptor/internal/datastore"
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
		Repo: repo.NewRKEClusterRepo(datastore.NewDB()),
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
