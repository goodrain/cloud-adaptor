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

package cmds

import (
	"fmt"
	"strconv"
	"time"

	cli "github.com/urfave/cli/v2"
	"goodrain.com/cloud-adaptor/adaptor"
	"goodrain.com/cloud-adaptor/adaptor/factory"
	"goodrain.com/cloud-adaptor/adaptor/v1alpha1"
	"goodrain.com/cloud-adaptor/operator"
	"goodrain.com/cloud-adaptor/util"
)

var clusterCommand = &cli.Command{
	Name: "cluster",
	Subcommands: []*cli.Command{
		&cli.Command{
			Name:    "list",
			Aliases: []string{"l"},
			Flags:   defatltFlag,
			Action:  listCluster,
			Usage:   "list all cluster",
		},
		&cli.Command{
			Name:   "get",
			Flags:  defatltFlag,
			Action: getCluster,
			Usage:  "show one cluster info, `get <CLUSTER_ID>`",
		},
		&cli.Command{
			Name:    "create",
			Aliases: []string{"c"},
			Flags: append(defatltFlag, &cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Value:   "rainbond-default-ack",
				Usage:   "the cluster name, if not privide,",
			}, &cli.StringFlag{
				Name:     "regionID",
				Aliases:  []string{"r"},
				Required: true,
				Usage:    "the region id for create cluster",
			}, &cli.IntFlag{
				Name:    "worker",
				Aliases: []string{"w"},
				Value:   2,
				Usage:   "the cluster worker node number",
			}, &cli.StringFlag{
				Name:    "workerType",
				Aliases: []string{"wt"},
				Usage:   "the cluster worker node instance type. eg. ecs.g6.16xlarge, if not provide, will select 16G/4Core",
			}, &cli.StringFlag{
				Name:    "vpcID",
				Aliases: []string{"vpc"},
				Usage:   "the id of the vpc to be used, if not provide, will create a new",
			}, &cli.StringFlag{
				Name:    "vswitchID",
				Aliases: []string{"vswitch"},
				Usage:   "the id of the vswitch to be used, if not provide, will create a new",
			}, &cli.StringFlag{
				Name:    "zoneID",
				Aliases: []string{"zone"},
				Usage:   "the id of the zone to be used, if not provide, will use first available zone in region",
			}),
			Action: createCluster,
			Usage:  "create a cluster",
		},
		&cli.Command{
			Name:   "region-config",
			Flags:  defatltFlag,
			Action: getClusterRegion,
			Usage:  "get rainbond region info in cluster, `region <CLUSTER_ID>`",
		},
		&cli.Command{
			Name:   "init-status",
			Flags:  defatltFlag,
			Action: getClusterRegionInitStatus,
			Usage:  "get rainbond region init status in cluster, `init-status <CLUSTER_ID>`",
		},
	},
}

func getAdaptor(ctx *cli.Context) (adaptor.CloudAdaptor, error) {
	secret := ctx.String("secret")
	if secret == "" {
		return nil, cli.Exit("secret must be set", 1)
	}
	accessKey := ctx.String("accessKey")
	if accessKey == "" {
		return nil, cli.Exit("accessKey must be set", 1)
	}
	adaptor := ctx.String("adaptor")
	if adaptor == "" {
		return nil, cli.Exit("adaptor must be set", 1)
	}
	return factory.GetCloudFactory().GetAdaptor(adaptor, accessKey, secret)
}

func listCluster(ctx *cli.Context) error {
	adaptor, err := getAdaptor(ctx)
	if err != nil {
		return err
	}
	clusters, err := adaptor.ClusterList()
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	table := util.NewTable(nil, nil)
	table.SetHeader([]string{"ID", "Name", "State", "Version", "VPCID", "VSwitchID", "ZoneID"})
	for _, cluster := range clusters {
		table.AddRow([]string{cluster.ClusterID, cluster.Name, cluster.State, cluster.CurrentVersion, cluster.VPCID, cluster.VSwitchID, cluster.ZoneID})
	}
	fmt.Println(table.Render())
	return nil
}

func getCluster(ctx *cli.Context) error {
	adaptor, err := getAdaptor(ctx)
	if err != nil {
		return err
	}
	clusterID := ctx.Args().First()
	if clusterID == "" {
		return cli.Exit("clusterID must be set, eg cluster <CLUSTER_ID>", 1)
	}
	cluster, err := adaptor.DescribeCluster(clusterID)
	if err != nil {
		return cli.Exit(err, 1)
	}
	yamlShow(cluster)
	return nil
}

func createCluster(ctx *cli.Context) error {
	adaptor, err := getAdaptor(ctx)
	if err != nil {
		return err
	}
	regionID := ctx.String("regionID")
	vpcID := ctx.String("vpcID")
	vswitchID := ctx.String("vswitchID")
	var rollback = []func(){}
	defer func() {
		for _, f := range rollback {
			f()
		}
	}()
	if vpcID == "" || vswitchID == "" {
		fmt.Printf("start create a new vpc %s \n", "rainbond-default-vpc")
		zoneID := ctx.String("zoneID")
		if zoneID == "" {
			zones, err := adaptor.ListZones(regionID)
			if err != nil {
				return cli.Exit(err, 1)
			}
			if len(zones) > 0 {
				zoneID = zones[0].ZoneID
			} else {
				return cli.Exit("not found available zone in region "+zoneID, 1)
			}
		}
		vpc := &v1alpha1.VPC{
			RegionID:  regionID,
			VpcName:   "rainbond-default-vpc",
			CidrBlock: "10.0.0.0/8",
		}
		if err := adaptor.CreateVPC(vpc); err != nil {
			return cli.Exit(err, 1)
		}
		rollback = append(rollback, func() {
			if err := adaptor.DeleteVPC(regionID, vpc.VpcID); err != nil {
				fmt.Printf("rollback vpc %s failure %s \n", vpc.VpcID, err.Error())
			} else {
				fmt.Println("rollback vpc success")
			}
		})
		retry := 0
		for retry < 3 {
			time.Sleep(time.Second * 2)
			vpcs, err := adaptor.VPCList(regionID)
			if err != nil {
				return cli.Exit(err, 1)
			}
			for _, vpc := range vpcs {
				if vpc.VpcName == "rainbond-default-vpc" && vpc.Status == "available" {
					break
				}
			}
			retry++
		}
		vpcID = vpc.VpcID
		vswitch := &v1alpha1.VSwitch{
			RegionID:    regionID,
			VpcID:       vpcID,
			CidrBlock:   "10.22.0.0/16",
			VSwitchName: "rainbond-default-vswitch",
			ZoneID:      zoneID,
		}
		if err := adaptor.CreateVSwitch(vswitch); err != nil {
			return cli.Exit(err, 1)
		}
		rollback = append([]func(){func() {
			if err := adaptor.DeleteVSwitch(regionID, vswitch.VSwitchID); err != nil {
				fmt.Printf("rollback vswitch failure %s \n", err.Error())
			} else {
				fmt.Println("rollback vswitch success")
			}

		}}, rollback...)
		vswitchID = vswitch.VSwitchID
		fmt.Printf("create a new vpc %s success\n", "rainbond-default-vpc")
	}
	name := ctx.String("name")
	fmt.Printf("start create a new ack cluster %s \n", name)

	var instanceType = ctx.String("workerType")
	if instanceType == "" {
		types, err := adaptor.ListInstanceType(regionID)
		if err != nil {
			return cli.Exit(err, 1)
		}
		for _, t := range types {
			if t.InstanceTypeID == "ecs.g6.xlarge" {
				instanceType = t.InstanceTypeID
				fmt.Println("select instance type is " + t.InstanceTypeID)
				break
			}
		}
		if instanceType == "" {
			for _, t := range types {
				if t.MemorySize == 16 && t.CPUCoreCount == 4 {
					instanceType = t.InstanceTypeID
					fmt.Println("select instance type is " + t.InstanceTypeID)
					break
				}
			}
		}
	}
	if instanceType == "" {
		return cli.Exit("do not select instance type, please provide by --workerType", 1)
	}
	clusterConfig := v1alpha1.GetDefaultACKCreateClusterConfig(v1alpha1.KubernetesClusterConfig{
		ClusterName:   name,
		Region:        regionID,
		WorkerNodeNum: ctx.Int("worker"),
		InstanceType:  instanceType,
		VpcID:         vpcID,
		VSwitchID:     vswitchID,
	})
	_, err = adaptor.CreateCluster(clusterConfig)
	if err != nil {
		return cli.Exit(err, 1)
	}
	fmt.Printf("create a new ack cluster %s success\n", name)
	rollback = nil
	return nil
}

func getClusterRegion(ctx *cli.Context) error {
	adaptor, err := getAdaptor(ctx)
	if err != nil {
		return err
	}
	clusterID := ctx.Args().First()
	if clusterID == "" {
		return cli.Exit("clusterID must be set, eg region <CLUSTER_ID>", 1)
	}
	kubeConfig, err := adaptor.GetKubeConfig(clusterID)
	if err != nil {
		return cli.Exit(err, 1)
	}
	rri := operator.NewRainbondRegionInit(*kubeConfig)
	status, err := rri.GetRainbondRegionStatus(clusterID)
	if err != nil {
		return cli.Exit(err, 1)
	}
	if status.RegionConfig != nil && status.RegionConfig.Data["apiAddress"] != "" {
		configMap := status.RegionConfig
		regionConfig := map[string]string{
			"client.pem":          string(configMap.BinaryData["client.pem"]),
			"client.key.pem":      string(configMap.BinaryData["client.key.pem"]),
			"ca.pem":              string(configMap.BinaryData["ca.pem"]),
			"apiAddress":          configMap.Data["apiAddress"],
			"websocketAddress":    configMap.Data["websocketAddress"],
			"defaultDomainSuffix": configMap.Data["defaultDomainSuffix"],
			"defaultTCPHost":      configMap.Data["defaultTCPHost"],
		}
		yamlShow(regionConfig)
	} else {
		fmt.Println("region config can not found, maybe not yet")
	}
	return nil
}

func getClusterRegionInitStatus(ctx *cli.Context) error {
	adaptor, err := getAdaptor(ctx)
	if err != nil {
		return err
	}
	clusterID := ctx.Args().First()
	if clusterID == "" {
		return cli.Exit("clusterID must be set, eg region <CLUSTER_ID>", 1)
	}
	kubeConfig, err := adaptor.GetKubeConfig(clusterID)
	if err != nil {
		return cli.Exit(err, 1)
	}
	rri := operator.NewRainbondRegionInit(*kubeConfig)
	status, err := rri.GetRainbondRegionStatus(clusterID)
	if err != nil {
		return cli.Exit(err, 1)
	}
	if status.OperatorReady {
		fmt.Printf("Operator is ready\n")
	}
	if status.RainbondVolume != nil && len(status.RainbondVolume.Status.Conditions) > 0 {
		for _, con := range status.RainbondVolume.Status.Conditions {
			if string(con.Type) == "Ready" && string(con.Status) == "True" {
				fmt.Printf("Rainbond volume is ready\n")
			}
		}
	}
	if status.RainbondCluster != nil {
		if status.RainbondCluster.Spec.ImageHub != nil && status.RainbondCluster.Spec.ImageHub.Domain != "" {
			fmt.Printf("Local image hub %s ready\n", status.RainbondCluster.Spec.ImageHub.Domain)
		} else {
			fmt.Printf("Local image hub not ready\n")
		}
	}
	if status.RainbondPackage != nil {
		table := util.NewTable(nil, nil)
		table.SetHeader([]string{"Condition", "Status", "Progress"})
		for _, con := range status.RainbondPackage.Status.Conditions {
			table.AddRow([]string{string(con.Type), string(con.Status), strconv.Itoa(con.Progress)})
		}
		fmt.Println(table.Render())
	}
	return nil
}
