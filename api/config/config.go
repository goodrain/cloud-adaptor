package config

import (
	"os"

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
	Addr string
	Port int
	User string
	Pass string
	Name string
}

//GetDefaultConfig get default config
func GetDefaultConfig(ctx *cli.Context) *Config {
	return &Config{
		TestMode: ctx.Bool("testMode"),
		LogLevel: ctx.String("logLevel"),
		NSQConfig: &NSQConfig{
			NsqLookupdAddress: ctx.String("nsq-lookupd-server"),
			NsqdAddress:       ctx.String("nsqd-server"),
		},
		DB: &DB{
			Addr: ctx.String("dbAddr"),
			Port: ctx.Int("dbPort"),
			User: ctx.String("dbUser"),
			Pass: ctx.String("dbPass"),
			Name: ctx.String("dbName"),
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
