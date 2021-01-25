package tke

import (
	"testing"
)

var testAccess = ""
var testSecret = ""

func TestDescribeCluster(t *testing.T) {
	adaptor, err := Create(testAccess, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	clusters, err := adaptor.ClusterList()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(clusters)
}
