package custom

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	rainbondv1alpha1 "github.com/goodrain/rainbond-operator/api/v1alpha1"
	"github.com/sirupsen/logrus"
	"goodrain.com/cloud-adaptor/adaptor"
	"goodrain.com/cloud-adaptor/adaptor/v1alpha1"
	"goodrain.com/cloud-adaptor/api/infrastructure/datastore"
	"goodrain.com/cloud-adaptor/api/models"
	"goodrain.com/cloud-adaptor/library/bcode"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type customAdaptor struct {
	Repo *ClusterRepo
}

//Create create ack adaptor
func Create() (adaptor.RainbondClusterAdaptor, error) {
	return &customAdaptor{
		Repo: NewCustomClusterRepo(datastore.GetGDB()),
	}, nil
}

func (c *customAdaptor) ClusterList() ([]*v1alpha1.Cluster, error) {
	clusters, err := c.Repo.ListCluster()
	if err != nil {
		return nil, err
	}
	var re []*v1alpha1.Cluster
	var wait sync.WaitGroup
	for _, clu := range clusters {
		wait.Add(1)
		go func(clu *models.CustomCluster) {
			defer wait.Done()
			cluster, err := c.DescribeCluster(clu.ClusterID)
			if err != nil {
				logrus.Warningf("query kubernetes cluster failure %s", err.Error())
			}
			if cluster != nil {
				re = append(re, cluster)
			}
		}(clu)
	}
	wait.Wait()
	return re, nil
}

func (c *customAdaptor) DescribeCluster(clusterID string) (*v1alpha1.Cluster, error) {
	cc, err := c.Repo.GetCluster(clusterID)
	if err != nil {
		return nil, fmt.Errorf("query cluster meta info failure %s", err.Error())
	}
	cluster := &v1alpha1.Cluster{
		Name:        cc.Name,
		ClusterID:   cc.ClusterID,
		Created:     v1alpha1.NewTime(cc.CreatedAt),
		State:       v1alpha1.OfflineState,
		ClusterType: "custom",
		EIP: func() []string {
			if cc.EIP != "" {
				return strings.Split(cc.EIP, ",")
			}
			return nil
		}(),
	}
	kc := v1alpha1.KubeConfig{Config: cc.KubeConfig}
	client, _, err := kc.GetKubeClient()
	if err != nil {
		return cluster, fmt.Errorf("create kube client failure %s", err.Error())
	}
	vinfo, err := client.ServerVersion()
	if err != nil {
		return cluster, fmt.Errorf("get cluster version failure  %s", err.Error())
	}
	cluster.KubernetesVersion = vinfo.String()
	cluster.CurrentVersion = vinfo.String()
	cluster.MasterURL.APIServerEndpoint, _ = kc.KubeServer()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	nodes, err := client.CoreV1().Nodes().List(ctx, v1.ListOptions{})
	if err != nil {
		cluster.Parameters["DisableRainbondInit"] = true
		cluster.Parameters["Message"] = "无法获取集群节点列表"
		return cluster, fmt.Errorf("query cluster node info failure %s", err.Error())
	}
	cluster.State = v1alpha1.RunningState
	cluster.Size = len(nodes.Items)
	_, err = client.CoreV1().ConfigMaps("rbd-system").Get(ctx, "region-config", v1.GetOptions{})
	if err == nil {
		cluster.RainbondInit = true
	}
	return cluster, nil
}

func (c *customAdaptor) GetKubeConfig(clusterID string) (*v1alpha1.KubeConfig, error) {
	cc, err := c.Repo.GetCluster(clusterID)
	if err != nil {
		return nil, fmt.Errorf("query cluster meta info failure %s", err.Error())
	}
	return &v1alpha1.KubeConfig{Config: cc.KubeConfig}, nil
}

//DeleteCluster delete cluster
func (c *customAdaptor) DeleteCluster(clusterID string) error {
	cluster, _ := c.DescribeCluster(clusterID)
	if cluster != nil && cluster.RainbondInit {
		return bcode.ErrClusterNotAllowDelete
	}
	return c.Repo.DeleteCluster(clusterID)
}

func (c *customAdaptor) GetRainbondInitConfig(cluster *v1alpha1.Cluster, gateway, chaos []*rainbondv1alpha1.K8sNode, rollback func(step, message, status string)) *v1alpha1.RainbondInitConfig {
	return &v1alpha1.RainbondInitConfig{
		EnableHA: func() bool {
			if cluster.Size > 3 {
				return true
			}
			return false
		}(),
		ClusterID:    cluster.ClusterID,
		GatewayNodes: gateway,
		ChaosNodes:   chaos,
		EIPs: func() (re []string) {
			if len(cluster.EIP) > 0 {
				return cluster.EIP
			}
			// Preferably use the IP address of KubeAPI as the EIP
			var kubeAPIHOST = ""
			url, _ := url.Parse(cluster.MasterURL.APIServerEndpoint)
			if url != nil {
				kubeAPIIP := net.ParseIP(url.Host)
				if kubeAPIIP != nil {
					kubeAPIHOST = url.Host
				}
			}
			for _, n := range gateway {
				if n.ExternalIP != "" {
					re = append(re, n.ExternalIP)
				}
				if kubeAPIHOST != "" && n.InternalIP == kubeAPIHOST {
					re = append(re, n.InternalIP)
				}
			}
			return
		}(),
	}
}

func (c *customAdaptor) CreateCluster(v1alpha1.CreateClusterConfig) (*v1alpha1.Cluster, error) {
	return nil, nil
}

func (c *customAdaptor) CreateRainbondKubernetes(ctx context.Context, config *v1alpha1.KubernetesClusterConfig, rollback func(step, message, status string)) *v1alpha1.Cluster {
	return nil
}

func (c *customAdaptor) ExpansionNode(ctx context.Context, en *v1alpha1.ExpansionNode, rollback func(step, message, status string)) *v1alpha1.Cluster {
	return nil
}
