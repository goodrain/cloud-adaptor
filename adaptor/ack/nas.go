package ack

import (
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/nas"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/sirupsen/logrus"
	"goodrain.com/cloud-adaptor/adaptor/v1alpha1"
)

func (a *ackAdaptor) GetNasZone(regionID string) (string, error) {
	client, err := nas.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return "", err
	}
	request := nas.CreateDescribeZonesRequest()
	request.Scheme = "https"
	request.RegionId = regionID
	response, err := client.DescribeZones(request)
	if err != nil {
		return "", err
	}
	fmt.Println(response.Zones.Zone)
	for _, zone := range response.Zones.Zone {
		for _, p := range zone.Capacity.Protocol {
			if p == "nfs" {
				return zone.ZoneId, nil
			}
		}
	}
	return "", fmt.Errorf("not found support Capacity nas zone")
}

//CreateNAS create nas, if zone is not found, will retry other zone
func (a *ackAdaptor) CreateNAS(clusterID, regionID, zoneID string) (string, error) {
	logrus.Infof("create nas in region %s zone %s", regionID, zoneID)
	client, err := nas.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return "", err
	}
	drequest := nas.CreateDescribeFileSystemsRequest()
	drequest.RegionId = regionID
	dresponse, err := client.DescribeFileSystems(drequest)
	if err != nil && !strings.Contains(err.Error(), "The specified resource does not exist") {
		return "", err
	}
	for _, system := range dresponse.FileSystems.FileSystem {
		if system.Description == "rainbond-region-nas_"+clusterID {
			logrus.Infof("nas filesystem for cluster %s is exist", clusterID)
			return system.FileSystemId, nil
		}
	}
	request := nas.CreateCreateFileSystemRequest()
	request.ProtocolType = "NFS"
	request.ZoneId = zoneID
	request.RegionId = regionID
	request.Description = "rainbond-region-nas_" + clusterID
	request.StorageType = "Capacity" //默认容量型，支持的region更多，性能型（Performance）支持的region有限
	request.Scheme = "https"
	response, createErr := client.CreateFileSystem(request)
	if createErr != nil {
		if real, ok := createErr.(*errors.ServerError); ok {
			if real.ErrorCode() == "InvalidAZone.NotFound" {
				logrus.Infof("because of %s, select nas zone in region %s ", real.Message(), regionID)
				// select other zone
				zone, err := a.GetNasZone(regionID)
				if err != nil {
					return "", err
				}
				//retry
				request := nas.CreateCreateFileSystemRequest()
				request.ZoneId = zone
				request.ProtocolType = "NFS"
				request.StorageType = "Capacity" //默认容量型，支持的region更多，性能型（Performance）支持的region有限
				request.Scheme = "https"
				zoneID = zone
				response, createErr = client.CreateFileSystem(request)
			}
		}
		if createErr != nil {
			return "", fmt.Errorf("create nas in region %s zone %s failure %s", regionID, request.ZoneId, createErr.Error())
		}
	}
	if !response.IsSuccess() {
		return "", fmt.Errorf("create nas in region %s zone %s failure:%s", regionID, zoneID, response.String())
	}
	// nas status is empty, do not check nas status
	return response.FileSystemId, nil
	// ticker := time.NewTicker(time.Second * 3)
	// timer := time.NewTimer(time.Minute * 10)
	// defer timer.Stop()
	// defer ticker.Stop()
	// for {
	// 	select {
	// 	case <-ticker.C:
	// 	case <-timer.C:
	// 		return "", fmt.Errorf("create nas timeout")
	// 	}
	// 	req := nas.CreateDescribeFileSystemsRequest()
	// 	req.FileSystemId = response.FileSystemId
	// 	res, err := client.DescribeFileSystems(req)
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	for _, nas := range res.FileSystems.FileSystem {
	// 		if nas.Status == "Running" {
	// 			logrus.Infof("create nas %s in region %s zone %s success", response.FileSystemId, regionID, request.ZoneId)
	// 			return response.FileSystemId, nil
	// 		}
	// 		logrus.Infof("nas %s status is %s", response.FileSystemId, nas.Status)
	// 	}
	// }
}

func (a *ackAdaptor) CreateNASMountTarget(clusterID, regionID, fileSystemID, VpcID, VSwitchID string) (string, error) {
	client, err := nas.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return "", err
	}
	req := nas.CreateDescribeMountTargetsRequest()
	req.Scheme = "https"
	req.FileSystemId = fileSystemID
	res, err := client.DescribeMountTargets(req)
	if err != nil && !strings.Contains(err.Error(), "The specified resource does not exist") {
		logrus.Errorf("describe mount target failure %s", err.Error())
		return "", err
	}
	for _, target := range res.MountTargets.MountTarget {
		logrus.Infof("nas %s mount %s status is %s", fileSystemID, target.MountTargetDomain, target.Status)
		if target.Status == "Active" {
			return target.MountTargetDomain, nil
		}
	}
	request := nas.CreateCreateMountTargetRequest()
	request.Scheme = "https"
	request.AccessGroupName = "DEFAULT_VPC_GROUP_NAME"
	request.FileSystemId = fileSystemID
	request.NetworkType = "VPC"
	request.VpcId = VpcID
	request.VSwitchId = VSwitchID
	response, err := client.CreateMountTarget(request)
	if err != nil {
		return "", fmt.Errorf("create nas mount target failure %s", err.Error())
	}
	if !response.IsSuccess() {
		return "", fmt.Errorf("create nas mount target failure:%s", response.String())
	}
	ticker := time.NewTicker(time.Second * 3)
	timer := time.NewTimer(time.Minute * 5)
	defer timer.Stop()
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-timer.C:
			return "", fmt.Errorf("create nas timeout")
		}
		req := nas.CreateDescribeMountTargetsRequest()
		req.Scheme = "https"
		req.FileSystemId = fileSystemID
		req.MountTargetDomain = response.MountTargetDomain
		res, err := client.DescribeMountTargets(req)
		if err != nil {
			logrus.Errorf("describe mount target failure %s", err.Error())
			continue
		}
		for _, target := range res.MountTargets.MountTarget {
			if target.Status == "Active" {
				return target.MountTargetDomain, nil
			}
			logrus.Infof("nas %s mount %s status is %s", fileSystemID, response.MountTargetDomain, target.Status)
		}
	}
}

func (a *ackAdaptor) GetNASInfo(regionID, fileSystemID string) (*v1alpha1.NasStorageInfo, error) {
	client, err := nas.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	request := nas.CreateDescribeFileSystemsRequest()
	request.RegionId = regionID
	request.FileSystemId = fileSystemID
	request.Scheme = "https"
	response, err := client.DescribeFileSystems(request)
	if err != nil {
		return nil, fmt.Errorf("describe nas failure %s", err.Error())
	}
	if !response.IsSuccess() {
		return nil, fmt.Errorf("describe nas failure:%s", response.String())
	}
	if len(response.FileSystems.FileSystem) < 1 {
		return nil, nil
	}
	nas := response.FileSystems.FileSystem[0]
	return &v1alpha1.NasStorageInfo{
		FileSystemID: nas.FileSystemId,
		Description:  nas.Description,
		CreateTime:   nas.CreateTime,
		RegionID:     nas.RegionId,
		ProtocolType: nas.ProtocolType,
		StorageType:  nas.StorageType,
		MeteredSize:  nas.MeteredSize,
		ZoneID:       nas.ZoneId,
		Bandwidth:    nas.Bandwidth,
		Capacity:     nas.Capacity,
		Status:       nas.Status,
	}, nil
}
