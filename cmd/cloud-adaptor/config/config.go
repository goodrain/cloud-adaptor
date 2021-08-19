// RAINBOND, Application Management Platform
// Copyright (C) 2020-2021 Goodrain Co., Ltd.

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

package config

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

// C represents a global configuration.
var C *Config

//Config config
type Config struct {
	TestMode  bool
	IsOffline bool
	LogLevel  string
	DB        *DB
	NSQConfig *NSQConfig
	Helm      *Helm
}

//NSQConfig config
type NSQConfig struct {
	NsqLookupdAddress string
	NsqdAddress       string
}

// DB holds configurations for database.
type DB struct {
	Host string
	Port int
	User string
	Pass string
	Name string
}

// Helm holds configurations for helm.
type Helm struct {
	RepoFile  string
	RepoCache string
}

func parseByEnvAndCtx(ctx *cli.Context, name, envName string) string {
	if os.Getenv(envName) != "" {
		return os.Getenv(envName)
	}
	return ctx.String(name)
}

func parseBoolByEnvAndCtx(ctx *cli.Context, name, envName string) bool {
	if os.Getenv(envName) != "" {
		parsed, err := strconv.ParseBool(os.Getenv(envName))
		if err == nil {
			return parsed
		}
	}
	return ctx.Bool(name)
}

func parseIntByEnvAndCtx(ctx *cli.Context, name, envName string) int {
	if os.Getenv(envName) != "" {
		parsed, err := strconv.Atoi(os.Getenv(envName))
		if err == nil {
			return parsed
		}
	}
	return ctx.Int(name)
}

//GetDefaultConfig get default config
func GetDefaultConfig(ctx *cli.Context) *Config {
	return &Config{
		TestMode:  parseBoolByEnvAndCtx(ctx, "testMode", "TEST_MODE"),
		IsOffline: parseBoolByEnvAndCtx(ctx, "isOffline", "IS_OFFLINE"),
		LogLevel:  parseByEnvAndCtx(ctx, "logLevel", "LOG_LEVEL"),
		NSQConfig: &NSQConfig{
			NsqLookupdAddress: parseByEnvAndCtx(ctx, "nsq-lookupd-server", "NSQ_LOOKUPD_SERVER"),
			NsqdAddress:       parseByEnvAndCtx(ctx, "nsqd-server", "NSQD_SERVER"),
		},
		DB: &DB{
			Host: parseByEnvAndCtx(ctx, "dbAddr", "MYSQL_HOST"),
			Port: parseIntByEnvAndCtx(ctx, "dbPort", "MYSQL_PORT"),
			User: parseByEnvAndCtx(ctx, "dbUser", "MYSQL_USER"),
			Pass: parseByEnvAndCtx(ctx, "dbPass", "MYSQL_PASS"),
			Name: parseByEnvAndCtx(ctx, "dbName", "MYSQL_DB"),
		},
		Helm: &Helm{
			RepoFile:  parseByEnvAndCtx(ctx, "helm-repo-file", "HELM_REPO_FILE"),
			RepoCache: parseByEnvAndCtx(ctx, "helm-cache", "HELM_CACHE"),
		},
	}
}

//Parse parse  command
func Parse(ctx *cli.Context) {
	c := GetDefaultConfig(ctx)
	C = c
}

// SetLogLevel Set log level
func SetLogLevel() {
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)
	switch C.LogLevel {
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}
