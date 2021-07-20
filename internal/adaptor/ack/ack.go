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
	"encoding/json"
	"fmt"
	"goodrain.com/cloud-adaptor/pkg/util/versionutil"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	rainbondv1alpha1 "github.com/goodrain/rainbond-operator/api/v1alpha1"
	"github.com/sirupsen/logrus"
	"goodrain.com/cloud-adaptor/internal/adaptor"
	"goodrain.com/cloud-adaptor/internal/adaptor/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
)

type ackAdaptor struct {
	//accessKeyID accessKey ID, it must have create cluster perm
	//https://help.aliyun.com/document_detail/86485.html?spm=a2c4g.11186623.2.11.196d1546DYhEjO#table-w5v-xxv-93g
	accessKeyID     string
	accessKeySecret string
	client          *sdk.Client
}

//Create create ack adaptor
func Create(accessKeyID, accessKeySecret string) (adaptor.CloudAdaptor, error) {
	client, err := sdk.NewClientWithAccessKey("", accessKeyID, accessKeySecret)
	if err != nil {
		return nil, err
	}
	return &ackAdaptor{
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		client:          client,
	}, nil
}

func (a *ackAdaptor) newRequest(method string) *requests.CommonRequest {
	request := requests.NewCommonRequest()
	request.Method = method
	request.Scheme = "https"
	request.Domain = "cs.aliyuncs.com"
	request.Version = "2015-12-15"
	request.Headers["Content-Type"] = "application/json"
	return request
}

func getInstanceType(set string) []string {
	if strings.Contains(set, ".xlarge") {
		return []string{
			set,
			"ecs.g5.xlarge",
			"ecs.g6.xlarge",
			"ecs.c6.xlarge",
		}
	}
	if strings.Contains(set, ".2xlarge") {
		return []string{
			set,
			"ecs.g5.2xlarge",
			"ecs.g6.2xlarge",
			"ecs.c6.2xlarge",
		}
	}
	if strings.Contains(set, ".3xlarge") {
		return []string{
			set,
			"ecs.g5.3xlarge",
			"ecs.g6.3xlarge",
			"ecs.c6.3xlarge",
		}
	}
	return []string{
		set,
		"ecs.g5.large",
		"ecs.g6.large",
	}
}

func (a *ackAdaptor) CreateRainbondKubernetes(ctx context.Context, eid string, config *v1alpha1.KubernetesClusterConfig, rollback func(step, message, status string)) *v1alpha1.Cluster {
	rollback("AllocateResource", "", "start")
	// select instance resource type
	//Resource type to be selected
	var selectInstanceType string
	var instanceTypes = getInstanceType(config.WorkerResourceType)
	var zoneID string
	for _, it := range instanceTypes {
		zones, err := a.DescribeAvailableResourceZones(config.Region, it)
		if err != nil {
			logrus.Errorf("list available zones failure %s", err.Error())
		}
		for _, z := range zones {
			if z.Status == "Available" {
				zoneID = z.ZoneID
				selectInstanceType = it
				break
			}
		}
		if selectInstanceType != "" && zoneID != "" {
			break
		}
	}
	if selectInstanceType == "" {
		rollback("AllocateResource", "Unable to find a suitable instance type, it may be that the region is currently sold out.", "failure")
		return nil
	}
	rollback("AllocateResource", selectInstanceType, "success")
	rollback("SelectZone", "", "start")
	// select zone
	rollback("SelectZone", zoneID, "success")
	if config.VpcID == "" {
		rollback("CreateVPC", "", "start")
		// create vpc
		vpc := &v1alpha1.VPC{
			RegionID:  config.Region,
			VpcName:   "rainbond-default-vpc",
			CidrBlock: "10.0.0.0/8",
		}
		if err := a.CreateVPC(vpc); err != nil {
			rollback("CreateVPC", err.Error(), "failure")
			return nil
		}
		rollback("CreateVPC", vpc.VpcID, "success")
		config.VpcID = vpc.VpcID
		rollback("CreateVSWitch", "", "start")
		// create vswitch
		vswitch := &v1alpha1.VSwitch{
			RegionID:    vpc.RegionID,
			VpcID:       vpc.VpcID,
			CidrBlock:   "10.22.0.0/16",
			VSwitchName: "rainbond-default-vswitch",
			ZoneID:      zoneID,
		}
		if err := a.CreateVSwitch(vswitch); err != nil {
			rollback("CreateVSwitch", err.Error(), "failure")
			return nil
		}
		rollback("CreateVSWitch", vswitch.VSwitchID, "success")
		config.VSwitchID = vswitch.VSwitchID
	}
	config.InstanceType = selectInstanceType
	clusterConfig := v1alpha1.GetDefaultACKCreateClusterConfig(*config)
	rollback("CreateCluster", "", "start")
	cluster, err := a.CreateCluster(eid, clusterConfig)
	if err != nil {
		rollback("CreateCluster", err.Error(), "failure")
		return nil
	}
	rollback("CreateCluster", cluster.ClusterID, "success")
	return cluster
}

func (a *ackAdaptor) ClusterList(eid string) ([]*v1alpha1.Cluster, error) {
	request := a.newRequest("GET")
	request.PathPattern = "/clusters"
	res, err := a.doRequest(request)
	if err != nil {
		return nil, fmt.Errorf("query cluster list from alibaba api failure %s", err.Error())
	}
	if !res.IsSuccess() {
		return nil, fmt.Errorf("query cluster list from alibaba api failure:%s", res.String())
	}
	var infos []*v1alpha1.Cluster
	if err := json.Unmarshal(res.GetHttpContentBytes(), &infos); err != nil {
		return nil, fmt.Errorf("unmarshal response failure:%s", err.Error())
	}
	var wait sync.WaitGroup
	for i := range infos {
		cluster := infos[i]
		if cluster.Parameters["ContainerCIDR"] != nil {
			cluster.PodCIDR = cluster.Parameters["ContainerCIDR"].(string)
		}
		if cluster.Parameters["DockerVersion"] != nil {
			cluster.DockerVersion = cluster.Parameters["DockerVersion"].(string)
		}
		if cluster.State == v1alpha1.InitState {
			cluster.CreateLogPath = fmt.Sprintf("https://cs.console.aliyun.com/#/k8s/cluster/%s/log", cluster.ClusterID)
		}
		wait.Add(1)
		go func(cluster *v1alpha1.Cluster) {
			defer wait.Done()
			cluster.Parameters = make(map[string]interface{})
			kube, _ := a.GetKubeConfig(eid, cluster.ClusterID)
			if kube != nil {
				coreclient, _, err := kube.GetKubeClient()
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				versionByte, err := coreclient.RESTClient().Get().AbsPath("/version").DoRaw(ctx)
				var info version.Info
				json.Unmarshal(versionByte, &info)
				if err == nil {
					cluster.CurrentVersion = info.String()
					if !versionutil.CheckVersion(cluster.CurrentVersion){
						cluster.Parameters["DisableRainbondInit"] = true
						cluster.Parameters["Message"] = fmt.Sprintf("当前集群版本为 %s ，无法继续初始化，初始化Rainbond支持的版本为1.16.x-1.19.x", cluster.CurrentVersion)
					}
				} else {
					cluster.Parameters["Message"] = "无法直接与集群 KubeAPI 通信"
					cluster.Parameters["DisableRainbondInit"] = true
				}
				_, err = coreclient.CoreV1().ConfigMaps("rbd-system").Get(ctx, "region-config", metav1.GetOptions{})
				if err == nil {
					cluster.RainbondInit = true
				}
			} else {
				cluster.Parameters["Message"] = "无法创建集群通信客户端"
				cluster.Parameters["DisableRainbondInit"] = true
			}
		}(cluster)
	}
	wait.Wait()
	return infos, nil
}

func (a *ackAdaptor) CreateCluster(eid string, config v1alpha1.CreateClusterConfig) (*v1alpha1.Cluster, error) {
	ackConfig, ok := config.(*v1alpha1.AckClusterConfig)
	if !ok {
		return nil, fmt.Errorf("config is valid")
	}
	body, err := json.Marshal(ackConfig)
	if err != nil {
		return nil, err
	}
	request := a.newRequest("POST")
	request.PathPattern = "/clusters"
	request.Content = body
	res, err := a.doRequest(request)
	if err != nil {
		return nil, fmt.Errorf("create ack cluster from alibaba api failure %s", err.Error())
	}
	if !res.IsSuccess() {
		return nil, fmt.Errorf("create ack cluster from alibaba api failure:%s", res.String())
	}
	var info v1alpha1.Cluster
	if err := json.Unmarshal(res.GetHttpContentBytes(), &info); err != nil {
		return nil, fmt.Errorf("unmarshal response failure:%s", err.Error())
	}
	return &info, nil
}

func (a *ackAdaptor) GetKubeConfig(eid string, clusterID string) (*v1alpha1.KubeConfig, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("cluster id can not be empty")
	}
	request := a.newRequest("GET")
	request.PathPattern = "/k8s/" + clusterID + "/user_config"
	res, err := a.doRequest(request)
	if err != nil {
		return nil, fmt.Errorf("query kube config from alibaba api failure %s", err.Error())
	}
	if !res.IsSuccess() {
		return nil, fmt.Errorf("query kube config from alibaba api failure:%s", res.String())
	}
	var infos v1alpha1.KubeConfig
	if err := json.Unmarshal(res.GetHttpContentBytes(), &infos); err != nil {
		return nil, fmt.Errorf("unmarshal response failure:%s", err.Error())
	}
	return &infos, nil
}

func (a *ackAdaptor) DescribeCluster(eid string, clusterID string) (*v1alpha1.Cluster, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("cluster id can not be empty")
	}
	request := a.newRequest("GET")
	request.PathPattern = "/clusters/" + clusterID
	res, err := a.doRequest(request)
	if err != nil {
		return nil, fmt.Errorf("query cluster info from alibaba api failure %s", err.Error())
	}
	if !res.IsSuccess() {
		return nil, fmt.Errorf("query cluster info from alibaba api failure:%s", res.String())
	}
	var info v1alpha1.Cluster
	if err := json.Unmarshal(res.GetHttpContentBytes(), &info); err != nil {
		return nil, fmt.Errorf("unmarshal response failure:%s", err.Error())
	}
	info.PodCIDR = info.Parameters["ContainerCIDR"].(string)
	info.DockerVersion = info.Parameters["DockerVersion"].(string)
	if info.State == v1alpha1.InitState {
		info.CreateLogPath = fmt.Sprintf("https://cs.console.aliyun.com/#/k8s/cluster/%s/log`;", clusterID)
	}

	return &info, nil
}

func (a *ackAdaptor) doRequest(request *requests.CommonRequest) (*responses.CommonResponse, error) {
	response, err := a.client.ProcessCommonRequest(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (a *ackAdaptor) VPCList(regionID string) ([]*v1alpha1.VPC, error) {
	vpcclient, err := vpc.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return nil, err
	}
	request := vpc.CreateDescribeVpcsRequest()
	request.Scheme = "https"
	request.PageSize = requests.NewInteger(50)
	response, err := vpcclient.DescribeVpcs(request)
	if err != nil {
		return nil, err
	}
	if !response.IsSuccess() {
		return nil, fmt.Errorf("query vpc list from alibaba api failure:%s", response.String())
	}
	var list []*v1alpha1.VPC
	for i := range response.Vpcs.Vpc {
		v := response.Vpcs.Vpc[i]
		list = append(list, a.vpcConver(v))
	}
	return list, nil
}

func (a *ackAdaptor) DescribeVPC(regionID, vpcID string) (*v1alpha1.VPC, error) {
	vpcclient, err := vpc.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return nil, err
	}
	request := vpc.CreateDescribeVpcsRequest()
	request.Scheme = "https"
	request.VpcId = vpcID
	request.PageSize = requests.NewInteger(50)
	response, err := vpcclient.DescribeVpcs(request)
	if err != nil {
		return nil, err
	}
	if !response.IsSuccess() {
		return nil, fmt.Errorf("query vpc list from alibaba api failure:%s", response.String())
	}
	// first one
	for i := range response.Vpcs.Vpc {
		return a.vpcConver(response.Vpcs.Vpc[i]), nil
	}
	return nil, fmt.Errorf("not found vpc")
}

func (a *ackAdaptor) vpcConver(v vpc.Vpc) *v1alpha1.VPC {
	return &v1alpha1.VPC{
		VpcID:           v.VpcId,
		RegionID:        v.RegionId,
		Status:          v.Status,
		VpcName:         v.VpcName,
		CreationTime:    v.CreationTime,
		CidrBlock:       v.CidrBlock,
		Ipv6CidrBlock:   v.Ipv6CidrBlock,
		VRouterID:       v.VRouterId,
		Description:     v.Description,
		IsDefault:       v.IsDefault,
		NetworkACLNum:   v.NetworkAclNum,
		ResourceGroupID: v.ResourceGroupId,
		CenStatus:       v.CenStatus,
	}
}

func (a *ackAdaptor) vswitchConver(v vpc.VSwitch) *v1alpha1.VSwitch {
	vs := &v1alpha1.VSwitch{
		VpcID:                   v.VpcId,
		VSwitchID:               v.VSwitchId,
		Status:                  v.Status,
		CidrBlock:               v.CidrBlock,
		Ipv6CidrBlock:           v.Ipv6CidrBlock,
		ZoneID:                  v.ZoneId,
		AvailableIPAddressCount: v.AvailableIpAddressCount,
		Description:             v.Description,
		VSwitchName:             v.VSwitchName,
		CreationTime:            v.CreationTime,
		IsDefault:               v.IsDefault,
		ResourceGroupID:         v.ResourceGroupId,
		NetworkACLID:            v.NetworkAclId,
	}
	var tags []v1alpha1.Tag
	for _, tag := range v.Tags.Tag {
		tags = append(tags, v1alpha1.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}
	vs.Tags = tags
	return vs
}

func (a *ackAdaptor) CreateVPC(v *v1alpha1.VPC) error {
	if v.RegionID == "" {
		return fmt.Errorf("not privide region id")
	}
	vpcclient, err := vpc.NewClientWithAccessKey(v.RegionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return err
	}
	request := vpc.CreateCreateVpcRequest()
	request.CidrBlock = v.CidrBlock
	request.Description = v.Description
	request.Ipv6CidrBlock = v.Ipv6CidrBlock
	request.EnableIpv6 = requests.NewBoolean(v.EnableIpv6)
	request.VpcName = v.VpcName
	request.ResourceGroupId = v.ResourceGroupID
	res, err := vpcclient.CreateVpc(request)
	if err != nil {
		return err
	}
	if !res.IsSuccess() {
		return fmt.Errorf("create vpc from alibaba api failure:%s", res.String())
	}
	v.VpcID = res.VpcId
	v.VRouterID = res.VRouterId
	v.ResourceGroupID = res.ResourceGroupId
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	timer := time.NewTimer(time.Minute * 5)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			return fmt.Errorf("create acount timeout")
		case <-ticker.C:
		}
		req := vpc.CreateDescribeVpcsRequest()
		req.Scheme = "https"
		req.VpcId = res.VpcId
		req.RegionId = v.RegionID
		res, err := vpcclient.DescribeVpcs(req)
		if err != nil {
			return err
		}
		for _, vpc := range res.Vpcs.Vpc {
			logrus.Infof("region %s vpc %s status is %s", v.RegionID, vpc.VpcId, vpc.Status)
			if vpc.Status == "Available" {
				return nil
			}
		}
	}
}

func (a *ackAdaptor) CreateVSwitch(v *v1alpha1.VSwitch) error {
	if v.RegionID == "" {
		return fmt.Errorf("not privide region id")
	}
	vpcclient, err := vpc.NewClientWithAccessKey(v.RegionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return err
	}
	request := vpc.CreateCreateVSwitchRequest()
	request.CidrBlock = v.CidrBlock
	request.Description = v.Description
	request.VpcId = v.VpcID
	request.ZoneId = v.ZoneID
	request.VSwitchName = v.VSwitchName
	request.Description = v.Description
	res, err := vpcclient.CreateVSwitch(request)
	if err != nil {
		return fmt.Errorf("create vswitch from alibaba api failure:%s", err.Error())
	}
	if !res.IsSuccess() {
		return fmt.Errorf("create vswitch from alibaba api failure:%s", res.String())
	}
	v.VSwitchID = res.VSwitchId
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	timer := time.NewTimer(time.Minute * 5)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			return fmt.Errorf("create acount timeout")
		case <-ticker.C:
		}
		req := vpc.CreateDescribeVSwitchesRequest()
		req.Scheme = "https"
		req.VSwitchId = res.VSwitchId
		req.RegionId = v.RegionID
		res, err := vpcclient.DescribeVSwitches(req)
		if err != nil {
			return err
		}
		for _, vs := range res.VSwitches.VSwitch {
			if vs.Status == "Available" {
				return nil
			}
			logrus.Infof("vs %s status is %s", vs.VSwitchId, vs.Status)
		}
	}
}

func (a *ackAdaptor) DescribeVSwitch(regionID, vswitchID string) (*v1alpha1.VSwitch, error) {
	vpcclient, err := vpc.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return nil, err
	}
	request := vpc.CreateDescribeVSwitchesRequest()
	request.Scheme = "https"
	request.VSwitchId = vswitchID
	request.PageSize = requests.NewInteger(50)
	response, err := vpcclient.DescribeVSwitches(request)
	if err != nil {
		return nil, err
	}
	if !response.IsSuccess() {
		return nil, fmt.Errorf("query vpc list from alibaba api failure:%s", response.String())
	}
	// first one
	for i := range response.VSwitches.VSwitch {
		return a.vswitchConver(response.VSwitches.VSwitch[i]), nil
	}
	return nil, fmt.Errorf("not found vswitch")
}

func (a *ackAdaptor) ListZones(regionID string) ([]*v1alpha1.Zone, error) {
	vpcclient, err := vpc.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return nil, err
	}
	request := vpc.CreateDescribeZonesRequest()
	request.Scheme = "https"
	response, err := vpcclient.DescribeZones(request)
	if err != nil {
		return nil, err
	}
	if !response.IsSuccess() {
		return nil, fmt.Errorf("query vpc list from alibaba api failure:%s", response.String())
	}
	var list []*v1alpha1.Zone
	for i := range response.Zones.Zone {
		v := response.Zones.Zone[i]
		list = append(list, &v1alpha1.Zone{
			ZoneID:    v.ZoneId,
			LocalName: v.LocalName,
		})
	}
	return list, nil
}

func (a *ackAdaptor) DeleteVPC(regionID, vpcID string) error {
	vpcclient, err := vpc.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return err
	}
	request := vpc.CreateDeleteVpcRequest()
	request.Scheme = "https"
	request.VpcId = vpcID
	response, err := vpcclient.DeleteVpc(request)
	if err != nil {
		if real, ok := err.(*errors.ServerError); ok {
			if real.ErrorCode() == "Forbbiden" {
				time.Sleep(time.Second * 1)
				response, err = vpcclient.DeleteVpc(request)
			}
		}
		if err != nil {
			return fmt.Errorf("delete vpc from alibaba api failure:%s", err.Error())
		}
	}
	if !response.IsSuccess() {
		return fmt.Errorf("delete vpc from alibaba api failure:%s", response.String())
	}
	return nil
}

func (a *ackAdaptor) DeleteVSwitch(regionID, vswitchID string) error {
	vpcclient, err := vpc.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return err
	}
	request := vpc.CreateDeleteVSwitchRequest()
	request.Scheme = "https"
	request.VSwitchId = vswitchID
	response, err := vpcclient.DeleteVSwitch(request)
	if err != nil {
		if real, ok := err.(*errors.ServerError); ok {
			if real.ErrorCode() == "IncorrectVSwitchStatus" {
				time.Sleep(time.Second * 1)
				response, err = vpcclient.DeleteVSwitch(request)
			}
		}
		if err != nil {
			return fmt.Errorf("delete vswitch from alibaba api failure:%s", err.Error())
		}
	}
	if !response.IsSuccess() {
		return fmt.Errorf("delete vswitch from alibaba api failure:%s", response.String())
	}
	return nil
}

func (a *ackAdaptor) ListInstanceType(regionID string) ([]*v1alpha1.InstanceType, error) {
	ecsclient, err := ecs.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return nil, err
	}
	request := ecs.CreateDescribeInstanceTypesRequest()
	request.Scheme = "https"
	response, err := ecsclient.DescribeInstanceTypes(request)
	if err != nil {
		return nil, fmt.Errorf("get instance types from alibaba api failure:%s", err.Error())
	}
	if !response.IsSuccess() {
		return nil, fmt.Errorf("delete vpc from alibaba api failure:%s", response.String())
	}
	var list []*v1alpha1.InstanceType
	for _, t := range response.InstanceTypes.InstanceType {
		list = append(list, &v1alpha1.InstanceType{
			InstanceTypeID:     t.InstanceTypeId,
			MemorySize:         t.MemorySize,
			CPUCoreCount:       t.CpuCoreCount,
			InstanceTypeFamily: t.InstanceTypeFamily,
		})
	}
	return list, nil
}

//GetECSIDByIPs get ecs id by vpcid and ips
func (a *ackAdaptor) GetECSIDByIPs(regionID, vpcID string, ips []string) (map[string]string, error) {
	client, err := ecs.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return nil, err
	}
	request := ecs.CreateDescribeInstancesRequest()
	request.Scheme = "https"
	request.RegionId = regionID
	request.VpcId = vpcID
	ipsBytes, _ := json.Marshal(ips)
	request.PrivateIpAddresses = string(ipsBytes)
	response, err := client.DescribeInstances(request)
	if err != nil {
		return nil, err
	}
	if !response.IsSuccess() {
		return nil, fmt.Errorf("get ecs id failure:%s", response.String())
	}
	var ids = make(map[string]string, len(ips))
	for _, instance := range response.Instances.Instance {
		if len(instance.VpcAttributes.PrivateIpAddress.IpAddress) > 0 {
			ids[instance.VpcAttributes.PrivateIpAddress.IpAddress[0]] = instance.InstanceId
		}
	}
	return ids, nil
}

// SetSecurityGroup set security rule
func (a *ackAdaptor) SetSecurityGroup(clusterID, regionID, securityGroupID string) error {
	client, err := ecs.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return err
	}
	request := ecs.CreateDescribeSecurityGroupAttributeRequest()
	request.Scheme = "https"
	request.SecurityGroupId = securityGroupID
	request.Direction = "ingress"
	response, err := client.DescribeSecurityGroupAttribute(request)
	if err != nil {
		return err
	}
	if !response.IsSuccess() {
		return fmt.Errorf("get security group attribute id failure:%s", response.String())
	}
	var createPerms = map[string]bool{
		"80/80":       false,
		"443/443":     false,
		"8443/8443":   false,
		"6060/6060":   false,
		"10000/11000": false,
	}
	for _, perm := range response.Permissions.Permission {
		_, exist := createPerms[perm.PortRange]
		if exist && perm.IpProtocol == "TCP" && perm.DestCidrIp == "0.0.0.0/0" {
			createPerms[perm.PortRange] = true
		}
	}
	for portRange, create := range createPerms {
		if !create {
			request := ecs.CreateAuthorizeSecurityGroupRequest()
			request.Scheme = "https"
			request.PortRange = portRange
			request.SecurityGroupId = securityGroupID
			request.IpProtocol = "tcp"
			request.SourceCidrIp = "0.0.0.0/0"
			presponse, perr := client.AuthorizeSecurityGroup(request)
			if perr != nil {
				presponse, perr = client.AuthorizeSecurityGroup(request)
				if perr != nil {
					logrus.Errorf("create security rule %s failure %s", portRange, perr.Error())
				}
				if !presponse.IsSuccess() {
					return fmt.Errorf("create security rule %s failure:%s", portRange, response.String())
				}
			}
		}
	}
	return nil
}

//DescribeAvailableResourceZones get support InstanceType zones
func (a *ackAdaptor) DescribeAvailableResourceZones(regionID, InstanceType string) ([]*v1alpha1.AvailableResourceZone, error) {
	client, err := ecs.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return nil, err
	}
	request := ecs.CreateDescribeAvailableResourceRequest()
	request.Scheme = "https"
	request.DestinationResource = "InstanceType"
	request.IoOptimized = "optimized"
	request.InstanceType = InstanceType
	response, err := client.DescribeAvailableResource(request)
	if err != nil {
		return nil, err
	}
	if !response.IsSuccess() {
		return nil, fmt.Errorf("get ecs id failure:%s", response.String())
	}
	var list []*v1alpha1.AvailableResourceZone
	for _, re := range response.AvailableZones.AvailableZone {
		list = append(list, &v1alpha1.AvailableResourceZone{
			Status:         re.Status,
			StatusCategory: re.StatusCategory,
			ZoneID:         re.ZoneId,
		})
	}
	return list, nil
}

//GetRainbondInitConfig get rainbond init config
func (a *ackAdaptor) GetRainbondInitConfig(eid string, cluster *v1alpha1.Cluster, gateway, chaos []*rainbondv1alpha1.K8sNode, rollback func(step, message, status string)) *v1alpha1.RainbondInitConfig {

	rollback("CreateRDS", "", "start")
	//指定pod cidr作为白名单
	regionDB := &v1alpha1.Database{
		Name:      "region",
		RegionID:  cluster.RegionID,
		UserName:  "rainbond_region",
		VPCID:     cluster.VPCID,
		ZoneID:    cluster.ZoneID,
		PodCIDR:   cluster.PodCIDR,
		VSwitchID: cluster.VSwitchID,
		Password:  cluster.ClusterID[0:16],
		ClusterID: cluster.ClusterID,
	}
	if err := a.CreateDB(regionDB); err != nil {
		rollback("CreateRDS", err.Error(), "failure")
		return nil
	}
	rollback("CreateRDS", regionDB.InstanceID, "success")
	// create nas
	vs, err := a.DescribeVSwitch(cluster.RegionID, cluster.VSwitchID)
	if err != nil {
		vs, err = a.DescribeVSwitch(cluster.RegionID, cluster.VSwitchID)
		if err != nil {
			rollback("CreateNAS", fmt.Sprintf("found vswitch %s with cluster failure %s", cluster.VSwitchID, err.Error()), "failure")
			return nil
		}
	}
	rollback("CreateNAS", "", "start")
	nasID, err := a.CreateNAS(cluster.ClusterID, cluster.RegionID, vs.ZoneID)
	if err != nil {
		rollback("CreateNAS", err.Error(), "failure")
		return nil
	}
	rollback("CreateNAS", nasID, "success")
	rollback("CreateNASMount", "", "start")
	nasMountDomain, err := a.CreateNASMountTarget(cluster.ClusterID, cluster.RegionID, nasID, cluster.VPCID, cluster.VSwitchID)
	if err != nil {
		rollback("CreateNASMount", err.Error(), "failure")
		return nil
	}
	rollback("CreateNASMount", nasMountDomain, "success")

	// create eip and bound
	rollback("CreateLoadBalancer", "", "start")
	slb, err := a.CreateLoadBalancer(cluster.ClusterID, cluster.RegionID)
	if err != nil {
		rollback("CreateLoadBalancer", err.Error(), "failure")
		return nil
	}
	rollback("CreateLoadBalancer", slb.LoadBalancerID+","+slb.Address, "success")

	// slb port 443 8443 80 6060 lb to cluster gateway node
	var gatewayIPs []string
	for _, g := range gateway {
		gatewayIPs = append(gatewayIPs, g.InternalIP)
	}
	rollback("BoundLoadBalancer", "", "start")
	logrus.Infof("gateway ips is %s", gatewayIPs)
	if err := a.BoundLoadBalancerToCluster(cluster.ClusterID, cluster.RegionID, cluster.VPCID, slb.LoadBalancerID, gatewayIPs); err != nil {
		rollback("BoundLoadBalancer", err.Error(), "failure")
		return nil
	}
	rollback("BoundLoadBalancer", "80,443,8443,6060", "success")

	// set security group
	rollback("SetSecurityGroup", "", "start")
	if err := a.SetSecurityGroup(cluster.ClusterID, cluster.RegionID, cluster.SecurityGroupID); err != nil {
		rollback("SetSecurityGroup", err.Error(), "failure")
	}
	rollback("SetSecurityGroup", "80/80,443/443,8443/8443,6060/6060,10000/11000", "success")
	return &v1alpha1.RainbondInitConfig{
		ClusterID:      cluster.ClusterID,
		RegionDatabase: regionDB,
		NasServer:      nasMountDomain,
		GatewayNodes:   gateway,
		ChaosNodes:     chaos,
		EIPs:           []string{slb.Address},
	}
}

//DeleteCluster delete cluster
func (a *ackAdaptor) DeleteCluster(eid string, clusterID string) error {
	return nil
}

func (a *ackAdaptor) ExpansionNode(ctx context.Context, eid string, en *v1alpha1.ExpansionNode, rollback func(step, message, status string)) *v1alpha1.Cluster {
	return nil
}
