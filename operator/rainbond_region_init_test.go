package operator

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/user"
	"path"
	"reflect"
	"testing"

	rainbondv1alpha1 "github.com/goodrain/rainbond-operator/api/v1alpha1"
	"goodrain.com/cloud-adaptor/adaptor/v1alpha1"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestInstall(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	configBytes, err := ioutil.ReadFile(path.Join(u.HomeDir, "/.kube/config"))
	if err != nil {
		t.Fatal(err)
	}
	rri := RainbondRegionInit{
		kubeconfig: v1alpha1.KubeConfig{Config: string(configBytes)},
	}
	if err := rri.InitRainbondRegion(&v1alpha1.RainbondInitConfig{
		EnableHA:          false,
		ClusterID:         "texxxxy",
		RainbondVersion:   "v5.3.0-cloud",
		RainbondCIVersion: "v5.3.0",
		SuffixHTTPHost:    "",
		GatewayNodes: []*rainbondv1alpha1.K8sNode{
			{Name: "192.168.56.104", InternalIP: "192.168.56.104"},
		},
		ChaosNodes: []*rainbondv1alpha1.K8sNode{
			{Name: "192.168.56.104", InternalIP: "192.168.56.104"},
		},
	}); err != nil {
		t.Fatal(err)
	}
}
func TestGetRainbondRegionStatus(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	configBytes, err := ioutil.ReadFile(path.Join(u.HomeDir, "/.kube/config"))
	if err != nil {
		t.Fatal(err)
	}
	rri := RainbondRegionInit{
		kubeconfig: v1alpha1.KubeConfig{Config: string(configBytes)},
	}
	status, err := rri.GetRainbondRegionStatus("")
	if err != nil {
		t.Fatal(err)
	}
	configMap := status.RegionConfig
	if configMap != nil {
		regionConfig := map[string]string{
			"client.pem":          string(configMap.BinaryData["client.pem"]),
			"client.key.pem":      string(configMap.BinaryData["client.key.pem"]),
			"ca.pem":              string(configMap.BinaryData["ca.pem"]),
			"apiAddress":          configMap.Data["apiAddress"],
			"websocketAddress":    configMap.Data["websocketAddress"],
			"defaultDomainSuffix": configMap.Data["defaultDomainSuffix"],
			"defaultTCPHost":      configMap.Data["defaultTCPHost"],
		}
		body, err := yaml.Marshal(regionConfig)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(string(body))
	}
	t.Logf("%+v", status)
}

func TestUninstallRegion(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	configBytes, err := ioutil.ReadFile(path.Join(u.HomeDir, "/.kube/config"))
	if err != nil {
		t.Fatal(err)
	}
	rri := RainbondRegionInit{
		kubeconfig: v1alpha1.KubeConfig{Config: string(configBytes)},
		namespace:  "rbd-system",
	}
	err = rri.UninstallRegion("")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	configBytes, err := ioutil.ReadFile(path.Join(u.HomeDir, "/.kube/config"))
	if err != nil {
		t.Fatal(err)
	}
	config := v1alpha1.KubeConfig{Config: string(configBytes)}
	_, baseclient, _ := config.GetKubeClient()
	var obj client.Object = &corev1.Pod{}
	var oldOjb = reflect.New(reflect.ValueOf(obj).Elem().Type()).Interface().(client.Object)
	if err := baseclient.Get(context.TODO(), types.NamespacedName{Name: "rainbond-operator-76b867cd66-5b7k4", Namespace: "rbd-system"}, oldOjb); err != nil {
		t.Fatal(err)
	}
	t.Log(oldOjb.GetName())
}
