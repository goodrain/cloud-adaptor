package ack

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/sirupsen/logrus"
	"goodrain.com/cloud-adaptor/internal/adaptor/v1alpha1"
)

func (a *ackAdaptor) CreateLoadBalancer(clusterID, regionID string) (*v1alpha1.LoadBalancer, error) {
	client, err := slb.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return nil, err
	}
	req := slb.CreateDescribeLoadBalancersRequest()
	req.Scheme = "https"
	req.RegionId = regionID
	res, err := client.DescribeLoadBalancers(req)
	if err != nil {
		return nil, err
	}
	for _, slb := range res.LoadBalancers.LoadBalancer {
		logrus.Infof("slb %s status is %s", slb.LoadBalancerId, slb.LoadBalancerStatus)
		if slb.LoadBalancerStatus == "active" && slb.LoadBalancerName == "rainbond-region-lb_"+clusterID {
			return a.slbConver(slb), nil
		}
	}
	request := slb.CreateCreateLoadBalancerRequest()
	request.Scheme = "https"
	request.RegionId = regionID
	request.AddressType = "internet"
	request.InternetChargeType = "paybytraffic"
	// 简约型实例，共享型实例已停售
	request.LoadBalancerSpec = "slb.s1.small"
	request.LoadBalancerName = "rainbond-region-lb_" + clusterID
	request.PayType = "PayOnDemand"
	response, err := client.CreateLoadBalancer(request)
	if err != nil {
		return nil, err
	}
	if !response.IsSuccess() {
		return nil, fmt.Errorf("create load balance failure:%s", response.String())
	}
	ticker := time.NewTicker(time.Second * 3)
	timer := time.NewTimer(time.Minute * 5)
	defer timer.Stop()
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-timer.C:
			return nil, fmt.Errorf("create slb timeout")
		}
		req := slb.CreateDescribeLoadBalancersRequest()
		req.Scheme = "https"
		req.RegionId = regionID
		req.LoadBalancerId = response.LoadBalancerId
		res, err := client.DescribeLoadBalancers(req)
		if err != nil {
			return nil, err
		}
		for _, slb := range res.LoadBalancers.LoadBalancer {
			if slb.LoadBalancerStatus == "active" {
				return a.slbConver(slb), nil
			}
			logrus.Infof("slb %s status is %s", response.LoadBalancerId, slb.LoadBalancerStatus)
		}
	}
}

func (a *ackAdaptor) slbConver(v slb.LoadBalancer) *v1alpha1.LoadBalancer {
	return &v1alpha1.LoadBalancer{
		LoadBalancerID:     v.LoadBalancerId,
		LoadBalancerName:   v.LoadBalancerName,
		LoadBalancerStatus: v.LoadBalancerStatus,
		Address:            v.Address,
		AddressType:        v.AddressType,
		RegionID:           v.RegionId,
		RegionIDAlias:      v.RegionIdAlias,
		VSwitchID:          v.VSwitchId,
		VpcID:              v.VpcId,
		NetworkType:        v.NetworkType,
		MasterZoneID:       v.MasterZoneId,
		SlaveZoneID:        v.SlaveZoneId,
		InternetChargeType: v.InternetChargeType,
		CreateTime:         v.CreateTime,
		CreateTimeStamp:    v.CreateTimeStamp,
		PayType:            v.PayType,
		ResourceGroupID:    v.ResourceGroupId,
		AddressIPVersion:   v.AddressIPVersion,
	}
}

func (a *ackAdaptor) createVServerGroup(clusterID, regionID, vpcID, loadBalancerID string, endpoints []string, port int) (string, error) {
	client, err := slb.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return "", err
	}
	hrequest := slb.CreateDescribeVServerGroupsRequest()
	hrequest.LoadBalancerId = loadBalancerID
	hresponse, err := client.DescribeVServerGroups(hrequest)
	if err != nil && !strings.Contains(err.Error(), "The specified resource does not exist") {
		return "", err
	}
	if hresponse != nil {
		for _, s := range hresponse.VServerGroups.VServerGroup {
			if s.VServerGroupName == fmt.Sprintf("rainbond-gateway-nodes-%d", port) {
				logrus.Infof("VServerGroupName is exist for cluster %s", clusterID)
				return s.VServerGroupId, nil
			}
		}
	}
	request := slb.CreateCreateVServerGroupRequest()
	request.Scheme = "https"
	request.LoadBalancerId = loadBalancerID
	request.VServerGroupName = fmt.Sprintf("rainbond-gateway-nodes-%d", port)
	ids, err := a.GetECSIDByIPs(regionID, vpcID, endpoints)
	if err != nil {
		return "", err
	}
	var servers []map[string]interface{}
	for ip, id := range ids {
		if id != "" && ip != "" {
			servers = append(servers, map[string]interface{}{
				"ServerId": id,
				"Type":     "ecs",
				"ServerIp": ip,
				"Port":     port,
			})
		}
	}
	serversBytes, err := json.Marshal(servers)
	if err != nil {
		return "", err
	}
	request.BackendServers = string(serversBytes)
	response, err := client.CreateVServerGroup(request)
	if err != nil {
		return "", fmt.Errorf("create load balance VServerGroup failure:%s", err.Error())
	}
	if !response.IsSuccess() {
		return "", fmt.Errorf("create load balance VServerGroup failure:%s", response.String())
	}
	return response.VServerGroupId, nil
}

//BoundLoadBalancerToCluster bound 443 80 8443 6060 port to cluster
func (a *ackAdaptor) BoundLoadBalancerToCluster(clusterID, regionID, vpcID, loadBalancerID string, endpoints []string) error {
	listenPorts := []int{80, 443, 8443, 6060}
	for _, port := range listenPorts {
		verserGroupID, err := a.createVServerGroup(clusterID, regionID, vpcID, loadBalancerID, endpoints, port)
		if err != nil {
			return err
		}
		if err := a.CreateLoadBalancerTCPListener(clusterID, regionID, loadBalancerID, verserGroupID, port); err != nil {
			return err
		}
	}
	return nil
}

func (a *ackAdaptor) CreateLoadBalancerTCPListener(clusterID, regionID, loadBalancerID, verserGroupID string, listenPort int) error {
	client, err := slb.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return err
	}
	req := slb.CreateDescribeLoadBalancerTCPListenerAttributeRequest()
	req.Scheme = "https"
	req.RegionId = regionID
	req.LoadBalancerId = loadBalancerID
	req.ListenerPort = requests.NewInteger(listenPort)
	res, err := client.DescribeLoadBalancerTCPListenerAttribute(req)
	if err != nil && !strings.Contains(err.Error(), "The specified resource does not exist") {
		return err
	}
	if res != nil && res.Status == "running" {
		logrus.Infof("create and start slb %s listener port %d success", loadBalancerID, listenPort)
		return nil
	}

	request := slb.CreateCreateLoadBalancerTCPListenerRequest()
	request.Scheme = "https"
	request.ListenerPort = requests.NewInteger(listenPort)
	request.Bandwidth = requests.NewInteger(-1)
	request.LoadBalancerId = loadBalancerID
	request.BackendServerPort = requests.NewInteger(listenPort)
	request.VServerGroupId = verserGroupID
	response, err := client.CreateLoadBalancerTCPListener(request)
	if err != nil {
		return err
	}
	if !response.IsSuccess() {
		return fmt.Errorf("create load balance %s tcp listener port %d failure:%s", loadBalancerID, listenPort, response.String())
	}
	// start listener
	srequest := slb.CreateStartLoadBalancerListenerRequest()
	srequest.Scheme = "https"
	srequest.ListenerPort = requests.NewInteger(listenPort)
	srequest.LoadBalancerId = loadBalancerID
	sresponse, err := client.StartLoadBalancerListener(srequest)
	if err != nil {
		return fmt.Errorf("start load balance %s tcp listenner port %d failure %s", loadBalancerID, listenPort, err.Error())
	}
	if !sresponse.IsSuccess() {
		return fmt.Errorf("start load balance %s tcp listener port %d failure:%s", loadBalancerID, listenPort, response.String())
	}
	// check listener status is running
	ticker := time.NewTicker(time.Second * 3)
	timer := time.NewTimer(time.Minute * 5)
	defer timer.Stop()
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-timer.C:
			return fmt.Errorf("create slb timeout")
		}
		req := slb.CreateDescribeLoadBalancerTCPListenerAttributeRequest()
		req.Scheme = "https"
		req.RegionId = regionID
		req.LoadBalancerId = loadBalancerID
		req.ListenerPort = requests.NewInteger(listenPort)
		res, err := client.DescribeLoadBalancerTCPListenerAttribute(req)
		if err != nil {
			return err
		}
		if res.Status == "running" {
			logrus.Infof("create and start slb %s listener port %d success", loadBalancerID, listenPort)
			return nil
		}
		logrus.Infof("slb %s listener port %d status is %s", loadBalancerID, listenPort, res.Status)
	}
}
