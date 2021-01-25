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
	yaml "gopkg.in/yaml.v2"
)

var commands cli.Commands

func init() {
	RegistClient(clusterCommand)
	RegistClient(vpcCommand)
	RegistClient(dbCommand)
	RegistClient(serverCommand)
	RegistClient(daemonCommand)
	RegistClient(dataCommand)
}

//GetClientCmds get client command
func GetClientCmds() cli.Commands {
	return commands
}

//RegistClient regist command
func RegistClient(c *cli.Command) {
	commands = append(commands, c)
}

var defatltFlag = []cli.Flag{
	&cli.StringFlag{
		Name:     "accessKey, k",
		Value:    "",
		EnvVars:  []string{"ACCESS_KEY"},
		Usage:    "the cloud user accesskey",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "secret, s",
		Value:    "",
		EnvVars:  []string{"ACCESS_SECRET"},
		Usage:    "the cloud user accesskey secret",
		Required: true,
	},
	&cli.StringFlag{
		Name:    "adaptor",
		Value:   "ack",
		EnvVars: []string{"ADAPTOR"},
		Usage:   "the cloud adaptor",
	}}

var dbInfoFlag = []cli.Flag{
	&cli.StringFlag{
		Name:    "dbAddr",
		Value:   "127.0.0.1",
		Usage:   "The address for database.",
		EnvVars: []string{"DB_ADDR"},
	},
	&cli.IntFlag{
		Name:    "dbPort",
		Value:   3306,
		Usage:   "The port for database.",
		EnvVars: []string{"DB_PORT"},
	},
	&cli.StringFlag{
		Name:    "dbUser",
		Value:   "root",
		Usage:   "The user for database.",
		EnvVars: []string{"DB_USER"},
	},
	&cli.StringFlag{
		Name:    "dbPass",
		Value:   "123456",
		Usage:   "The password for database.",
		EnvVars: []string{"DB_PASS"},
	},
	&cli.StringFlag{
		Name:    "dbName",
		Value:   "console",
		Usage:   "The name for database.",
		EnvVars: []string{"DB_NAME"},
	},
}

func yamlShow(item interface{}) {
	o, _ := yaml.Marshal(item)
	fmt.Println(string(o))
}
