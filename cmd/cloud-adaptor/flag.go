package main

import cli "github.com/urfave/cli/v2"

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
