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
	"testing"

	"goodrain.com/cloud-adaptor/internal/adaptor/v1alpha1"
)

func TestCreateKubernetesCluster(t *testing.T) {
	task, err := CreateTask(CreateKubernetesTask, &v1alpha1.KubernetesClusterConfig{
		ClusterName:        "rainbond-cluster",
		WorkerResourceType: "ecs.g5.large",
		WorkerNodeNum:      2,
		Provider:           "ack",
		Region:             "cn-huhehaote",
		SecretKey:          "hBsW4mlp35xQlqvqvm5Izmbt2UFR6E",
		AccessKey:          "LTAI4FtxHG8A8h328zBBNMtw",
	})
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			message := <-task.GetChan()
			fmt.Println(message)
		}
	}()
	task.Run(context.TODO())
}

func TestRKECreateKubernetesCluster(t *testing.T) {
	task, err := CreateTask(CreateKubernetesTask, &v1alpha1.KubernetesClusterConfig{
		ClusterName: "rainbond-cluster",
		Nodes: []v1alpha1.ConfigNode{
			{
				IP:      "192.168.56.104",
				SSHUser: "docker",
				SSHPort: 22,
				Roles:   []string{"etcd", "controlplane", "worker"},
			},
			{
				IP:      "192.168.56.103",
				SSHUser: "docker",
				SSHPort: 22,
				Roles:   []string{"worker"},
			},
		},
		Provider: "rke",
	})
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			message := <-task.GetChan()
			fmt.Println(message)
		}
	}()
	task.Run(context.TODO())
}
