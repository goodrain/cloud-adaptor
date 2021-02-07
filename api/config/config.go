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
	LogLevel  string
	DB        *DB
	NSQConfig *NSQConfig
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
		TestMode: parseBoolByEnvAndCtx(ctx, "testMode", "TEST_MODE"),
		LogLevel: parseByEnvAndCtx(ctx, "logLevel", "LOG_LEVEL"),
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
