// RAINBOND, Application Management Platform
// Copyright (C) 2014-2017 Goodrain Co., Ltd.

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

package ack

import (
	"context"
	"testing"

	rainbondv1alpha1 "github.com/goodrain/rainbond-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"goodrain.com/cloud-adaptor/adaptor/v1alpha1"
	"goodrain.com/cloud-adaptor/util"
	v1 "k8s.io/api/core/v1"
)

var testAccess = ""
var testSecret = ""

func TestListCluster(t *testing.T) {
	adaptor, err := Create(testAccess, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	list, err := adaptor.ClusterList("test")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range list {
		t.Logf("%+v", c)
	}
}

func TestGetCluster(t *testing.T) {
	adaptor, err := Create(testAccess, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	cluster, err := adaptor.DescribeCluster("test", "c528e9ce890cb4b9cbddb3f25c36bfd7d")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", cluster)
}

func TestGetKubeConfig(t *testing.T) {
	adaptor, err := Create(testAccess, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	config, err := adaptor.GetKubeConfig("test", "cd06fdbf66e974bf6a62a8dba27983523")
	if err != nil {
		t.Fatal(err)
	}
	config.Save("/tmp/c9b4516adec504a458bdf9a6891fc74fe.kubeconfig")
	coreClient, _, err := config.GetKubeClient()
	if err != nil {
		t.Fatal(err)
		return
	}
	nodes, err := coreClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(nodes)
}

func TestListVPC(t *testing.T) {
	adaptor, err := Create(testAccess, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	list, err := adaptor.VPCList("cn-huhehaote")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", list)
}

func TestCreateVPC(t *testing.T) {
	adaptor, err := Create(testAccess, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	vpc := &v1alpha1.VPC{
		RegionID: "cn-huhehaote",
		VpcName:  "rainbond-default-vpc",
	}
	if err := adaptor.CreateVPC(vpc); err != nil {
		t.Fatal(err)
	}
	t.Logf(vpc.VpcID)
}

func TestListInstanceType(t *testing.T) {
	adaptor, err := Create(testAccess, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	list, err := adaptor.ListInstanceType("cn-huhehaote")
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range list {
		t.Logf("%+v", l)
	}
}

func TestCreateDB(t *testing.T) {
	adaptor, err := Create(testAccess, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	db := &v1alpha1.Database{
		Name:     "console",
		RegionID: "cn-huhehaote",
		UserName: "console",
		Password: util.RandString(10),
	}
	if err := adaptor.CreateDB(db); err != nil {
		t.Fatal(err)
	}
	t.Logf("dh addr %s:%d", db.Host, db.Port)
}

func TestGetNasZone(t *testing.T) {
	a, err := Create(testAccess, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	adaptor := a.(*ackAdaptor)
	zone, err := adaptor.GetNasZone("cn-huhehaote")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(zone)
}

func TestCreateNAS(t *testing.T) {
	a, err := Create(testAccess, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	adaptor := a.(*ackAdaptor)
	zone, err := adaptor.CreateNAS("", "cn-hangzhou", "cn-hangzhou-a")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(zone)
}

func TestCreateNASMountTarget(t *testing.T) {
	a, err := Create(testAccess, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	adaptor := a.(*ackAdaptor)
	zone, err := adaptor.CreateNASMountTarget("", "cn-huhehaote", "8b4ce4ab50", "vpc-hp3tpsrybgmxndcra7c6c", "vsw-hp3yca62ng8wmbbsbnjsl")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(zone)
}

func TestCreateLoadBalancer(t *testing.T) {
	a, err := Create(testAccess, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	adaptor := a.(*ackAdaptor)
	lb, err := adaptor.CreateLoadBalancer("", "cn-huhehaote")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(lb.Address)
	t.Logf(lb.AddressType)
}

func TestBoundLoadBalancerToCluster(t *testing.T) {
	a, err := Create(testAccess, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	adaptor := a.(*ackAdaptor)
	err = adaptor.BoundLoadBalancerToCluster("", "cn-huhehaote", "vpc-hp3tpsrybgmxndcra7c6c", "lb-hp3pq3q70uoa173d3jbry", []string{"10.22.133.191", "10.22.133.192"})
	if err != nil {
		t.Fatal(err)
	}
}

func getK8sNode(node v1.Node) *rainbondv1alpha1.K8sNode {
	var Knode rainbondv1alpha1.K8sNode
	for _, address := range node.Status.Addresses {
		if address.Type == v1.NodeInternalIP {
			Knode.InternalIP = address.Address
		}
		if address.Type == v1.NodeExternalIP {
			Knode.ExternalIP = address.Address
		}
		if address.Type == v1.NodeHostName {
			Knode.Name = address.Address
		}
	}
	return &Knode
}

func TestSetSecurityGroup(t *testing.T) {
	a, err := Create(testAccess, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	adaptor := a.(*ackAdaptor)
	if err := adaptor.SetSecurityGroup("", "cn-huhehaote", "sg-hp3i7l3nngl8nqvw8tt9"); err != nil {
		t.Fatal(err)
	}
}

func TestDescribeAvailableResourceZones(t *testing.T) {
	a, err := Create(testAccess, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	adaptor := a.(*ackAdaptor)
	zones, err := adaptor.DescribeAvailableResourceZones("cn-hangzhou", "ecs.g5.large")
	if err != nil {
		t.Fatal(err)
	}
	for _, z := range zones {
		t.Logf("zone %s status %s", z.ZoneID, z.Status)
	}
}
