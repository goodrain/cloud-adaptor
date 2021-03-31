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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"time"

	rainbondv1alpha1 "github.com/goodrain/rainbond-operator/api/v1alpha1"
	"github.com/rancher/rke/k8s"
	"github.com/sirupsen/logrus"
	"goodrain.com/cloud-adaptor/adaptor/factory"
	typesv1 "goodrain.com/cloud-adaptor/api/openapi/types/v1"
	"goodrain.com/cloud-adaptor/operator"
	"goodrain.com/cloud-adaptor/version"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//InitRainbondConfig init rainbond region config
type InitRainbondConfig struct {
	EnterpriseID string `json:"eid"`
	ClusterID    string `json:"cluster_id"`
	AccessKey    string `json:"access_key"`
	SecretKey    string `json:"secret_key"`
	Provider     string `json:"provider"`
}

//InitRainbondCluster init rainbond cluster
type InitRainbondCluster struct {
	config *InitRainbondConfig
	result chan typesv1.Message
}

func (c *InitRainbondCluster) rollback(step, message, status string) {
	if status == "failure" {
		logrus.Errorf("%s failure, Message: %s", step, message)
	}
	c.result <- typesv1.Message{StepType: step, Message: message, Status: status}
}

//Run run take time 214.10s
func (c *InitRainbondCluster) Run(ctx context.Context) {
	defer c.rollback("Close", "", "")
	c.rollback("Init", "", "start")
	// create adaptor
	adaptor, err := factory.GetCloudFactory().GetRainbondClusterAdaptor(c.config.Provider, c.config.AccessKey, c.config.SecretKey)
	if err != nil {
		c.rollback("Init", fmt.Sprintf("create cloud adaptor failure %s", err.Error()), "failure")
		return
	}

	c.rollback("Init", "cloud adaptor create success", "success")
	c.rollback("CheckCluster", "", "start")
	// get kubernetes cluster info
	cluster, err := adaptor.DescribeCluster(c.config.EnterpriseID, c.config.ClusterID)
	if err != nil {
		cluster, err = adaptor.DescribeCluster(c.config.EnterpriseID, c.config.ClusterID)
		if err != nil {
			c.rollback("CheckCluster", err.Error(), "failure")
			return
		}
	}
	// check cluster status
	if cluster.State != "running" {
		c.rollback("CheckCluster", fmt.Sprintf("cluster status is %s,not support init rainbond", cluster.State), "failure")
		return
	}
	// check cluster connection status
	logrus.Infof("init kubernetes url %s", cluster.MasterURL)
	if cluster.MasterURL.APIServerEndpoint == "" {
		c.rollback("CheckCluster", fmt.Sprintf("cluster api not open eip,not support init rainbond"), "failure")
		return
	}

	kubeConfig, err := adaptor.GetKubeConfig(c.config.EnterpriseID, c.config.ClusterID)
	if err != nil {
		kubeConfig, err = adaptor.GetKubeConfig(c.config.EnterpriseID, c.config.ClusterID)
		if err != nil {
			c.rollback("CheckCluster", fmt.Sprintf("get kube config failure %s", err.Error()), "failure")
			return
		}
	}

	// check cluster not init rainbond
	coreClient, _, err := kubeConfig.GetKubeClient()
	if err != nil {
		c.rollback("CheckCluster", fmt.Sprintf("get kube config failure %s", err.Error()), "failure")
		return
	}

	// get cluster node lists
	getctx, cancel := context.WithTimeout(ctx, time.Second*10)
	nodes, err := coreClient.CoreV1().Nodes().List(getctx, metav1.ListOptions{})
	if err != nil {
		nodes, err = coreClient.CoreV1().Nodes().List(getctx, metav1.ListOptions{})
		cancel()
		if err != nil {
			logrus.Errorf("get kubernetes cluster node failure %s", err.Error())
			c.rollback("CheckCluster", "cluster node list can not found, please check cluster public access and account authorization", "failure")
			return
		}
	} else {
		cancel()
	}
	if len(nodes.Items) == 0 {
		c.rollback("CheckCluster", "node num is 0, can not init rainbond", "failure")
		return
	}
	c.rollback("CheckCluster", c.config.ClusterID, "success")

	// select gateway and chaos node
	gatewayNodes, chaosNodes := c.GetRainbondGatewayNodeAndChaosNodes(nodes.Items)
	initConfig := adaptor.GetRainbondInitConfig(c.config.EnterpriseID, cluster, gatewayNodes, chaosNodes, c.rollback)
	initConfig.RainbondVersion = version.RainbondRegionVersion
	// init rainbond
	c.rollback("InitRainbondRegionOperator", "", "start")
	if len(initConfig.EIPs) == 0 {
		c.rollback("InitRainbondRegionOperator", "can not select eip", "failure")
		return
	}
	rri := operator.NewRainbondRegionInit(*kubeConfig)
	if err := rri.InitRainbondRegion(initConfig); err != nil {
		c.rollback("InitRainbondRegionOperator", err.Error(), "failure")
		return
	}
	ticker := time.NewTicker(time.Second * 5)
	timer := time.NewTimer(time.Minute * 30)
	defer timer.Stop()
	defer ticker.Stop()
	var operatorMessage, imageHubMessage, packageMessage, apiReadyMessage bool
	for {
		select {
		case <-ctx.Done():
			c.rollback("InitRainbondRegion", "context cancel", "failure")
			return
		case <-ticker.C:
		case <-timer.C:
			c.rollback("InitRainbondRegion", "waiting rainbond region ready timeout", "failure")
			return
		}
		status, err := rri.GetRainbondRegionStatus(initConfig.ClusterID)
		if err != nil {
			if errors.IsNotFound(err) {
				c.rollback("InitRainbondRegion", err.Error(), "failure")
				return
			}
			logrus.Errorf("get rainbond region status failure %s", err.Error())
		}
		if status == nil {
			continue
		}
		if status.OperatorReady && !operatorMessage {
			c.rollback("InitRainbondRegionOperator", "", "success")
			c.rollback("InitRainbondRegionImageHub", "", "start")
			operatorMessage = true
			continue
		}

		if status.RainbondCluster.Spec.ImageHub != nil && status.RainbondCluster.Spec.ImageHub.Domain != "" && !imageHubMessage {
			c.rollback("InitRainbondRegionImageHub", "", "success")
			c.rollback("InitRainbondRegionPackage", "", "start")
			imageHubMessage = true
			continue
		}
		statusStr := fmt.Sprintf("Push Images:%d/%d\t", len(status.RainbondPackage.Status.ImagesPushed), status.RainbondPackage.Status.ImagesNumber)
		for _, con := range status.RainbondCluster.Status.Conditions {
			if con.Status == v1.ConditionTrue {
				statusStr += fmt.Sprintf("%s=>%s;\t", con.Type, con.Status)
			} else {
				statusStr += fmt.Sprintf("%s=>%s=>%s=>%s;\t", con.Type, con.Status, con.Reason, con.Message)
			}
		}
		logrus.Infof("cluster %s states: %s", cluster.Name, statusStr)

		for _, con := range status.RainbondPackage.Status.Conditions {
			if con.Type == rainbondv1alpha1.Ready && con.Status == rainbondv1alpha1.Completed && !packageMessage {
				c.rollback("InitRainbondRegionPackage", "", "success")
				c.rollback("InitRainbondRegionRegionConfig", "", "start")
				packageMessage = true
			}
			continue
		}

		if status.RegionConfig != nil && packageMessage {
			if checkAPIHealthy(status.RegionConfig) && !apiReadyMessage {
				c.rollback("InitRainbondRegionRegionConfig", "", "success")
				apiReadyMessage = true
				break
			}
		}
	}
	c.rollback("InitRainbondRegion", cluster.ClusterID, "success")
}

//GetRainbondGatewayNodeAndChaosNodes get gateway nodes
func (c *InitRainbondCluster) GetRainbondGatewayNodeAndChaosNodes(nodes []v1.Node) (gatewayNodes, chaosNodes []*rainbondv1alpha1.K8sNode) {
	for _, node := range nodes {
		if node.Annotations["rainbond.io/gateway-node"] == "true" {
			gatewayNodes = append(gatewayNodes, getK8sNode(node))
		}
		if node.Annotations["rainbond.io/chaos-node"] == "true" {
			chaosNodes = append(chaosNodes, getK8sNode(node))
		}
	}
	if len(gatewayNodes) == 0 {
		if len(nodes) < 2 {
			gatewayNodes = []*rainbondv1alpha1.K8sNode{
				getK8sNode(nodes[0]),
			}
		} else {
			gatewayNodes = []*rainbondv1alpha1.K8sNode{
				getK8sNode(nodes[0]),
				getK8sNode(nodes[1]),
			}
		}
	}
	if len(chaosNodes) == 0 {
		if len(nodes) < 2 {
			chaosNodes = []*rainbondv1alpha1.K8sNode{
				getK8sNode(nodes[0]),
			}
		} else {
			chaosNodes = []*rainbondv1alpha1.K8sNode{
				getK8sNode(nodes[0]),
				getK8sNode(nodes[1]),
			}
		}
	}
	return
}

// Stop init
func (c *InitRainbondCluster) Stop() error {
	return nil
}

//GetChan get message chan
func (c *InitRainbondCluster) GetChan() chan typesv1.Message {
	return c.result
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
	if externamAddress, exist := node.Annotations[k8s.ExternalAddressAnnotation]; exist && externamAddress != "" {
		logrus.Infof("set node %s externalIP %s by %s", node.Name, externamAddress, k8s.ExternalAddressAnnotation)
		Knode.ExternalIP = externamAddress
	}
	return &Knode
}

func checkAPIHealthy(configmap *v1.ConfigMap) bool {
	if configmap.BinaryData["ca.pem"] == nil {
		return false
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(configmap.BinaryData["ca.pem"])
	cliCrt, err := tls.X509KeyPair(configmap.BinaryData["client.pem"], configmap.BinaryData["client.key.pem"])
	if err != nil {
		logrus.Errorf("Loadx509keypair err: %s", err)
		return false
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      pool,
			Certificates: []tls.Certificate{cliCrt},
		},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}
	url := fmt.Sprintf("%s/v2/health", configmap.Data["apiAddress"])
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logrus.Errorf("create request failure: %s", err)
		return false
	}
	res, err := client.Do(req)
	if err != nil {
		logrus.Errorf("ping region api failure: %s", err)
		return false
	}
	if res.StatusCode == 200 {
		return true
	}
	return false
}
