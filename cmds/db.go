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
	"goodrain.com/cloud-adaptor/internal/adaptor/v1alpha1"
	"goodrain.com/cloud-adaptor/pkg/util"
)

var dbCommand = &cli.Command{
	Name: "database",
	Subcommands: []*cli.Command{
		{
			Name:    "create",
			Aliases: []string{"c"},
			Flags: append(defatltFlag, &cli.StringFlag{
				Name:     "regionID",
				Aliases:  []string{"r"},
				Required: true,
				Usage:    "the region id for create db",
			}, &cli.StringFlag{
				Name:    "dbname",
				Aliases: []string{"d"},
				Value:   "region",
				Usage:   "database name",
			}, &cli.StringFlag{
				Name:    "username",
				Aliases: []string{"u"},
				Value:   "region-user",
				Usage:   "database user name",
			}),
			Action: createDatabase,
			Usage:  "create a database for region",
		},
	},
}

func createDatabase(ctx *cli.Context) error {
	adaptor, err := getAdaptor(ctx)
	if err != nil {
		return err
	}
	regionID := ctx.String("regionID")
	db := &v1alpha1.Database{
		Name:     ctx.String("dbname"),
		RegionID: regionID,
		UserName: ctx.String("username"),
		Password: util.RandString(10),
	}
	err = adaptor.CreateDB(db)
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	table := util.NewTable(nil, nil)
	table.SetHeader([]string{"Address", "Database", "User", "Password"})
	table.AddRow([]string{fmt.Sprintf("%s:%d", db.Host, db.Port), db.Name, db.UserName, db.Password})
	fmt.Println(table.Render())
	return nil
}
