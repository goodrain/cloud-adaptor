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

	cli "github.com/urfave/cli/v2"
	"goodrain.com/cloud-adaptor/util"
)

var vpcCommand = &cli.Command{
	Name: "vpc",
	Subcommands: []*cli.Command{
		{
			Name:    "list",
			Aliases: []string{"l"},
			Flags: append(defatltFlag, &cli.StringFlag{
				Name:     "regionID",
				Aliases:  []string{"r"},
				Required: true,
				Usage:    "the region id for create cluster",
			}),
			Action: listVPC,
			Usage:  "list vpc in region",
		},
		{
			Name:   "get",
			Flags:  defatltFlag,
			Action: getVPC,
			Usage:  "show one cluster info, `get <CLUSTER_ID>`",
		},
		{
			Name:    "create",
			Aliases: []string{"c"},
			Flags: append(defatltFlag, &cli.StringFlag{
				Name:     "name",
				Aliases:  []string{"n"},
				Required: true,
				Usage:    "the cluster name",
			}, &cli.StringFlag{
				Name:     "regionID",
				Aliases:  []string{"r"},
				Required: true,
				Usage:    "the region id for create cluster",
			}),
			Action: createVPC,
			Usage:  "create a vpc",
		},
	},
}

func listVPC(ctx *cli.Context) error {
	adaptor, err := getAdaptor(ctx)
	if err != nil {
		return err
	}
	regionID := ctx.String("regionID")
	vpcs, err := adaptor.VPCList(regionID)
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	table := util.NewTable(nil, nil)
	table.SetHeader([]string{"ID", "Name", "VRouterID", "RegionID", "Status"})
	for _, vpc := range vpcs {
		table.AddRow([]string{vpc.VpcID, vpc.VpcName, vpc.VRouterID, vpc.RegionID, vpc.Status})
	}
	fmt.Println(table.Render())
	return nil
}

func getVPC(ctx *cli.Context) error {
	return nil
}

func createVPC(ctx *cli.Context) error {
	return nil
}
