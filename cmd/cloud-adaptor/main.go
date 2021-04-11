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

package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"goodrain.com/cloud-adaptor/cmd/cloud-adaptor/config"
	"goodrain.com/cloud-adaptor/internal/handler"
	"goodrain.com/cloud-adaptor/pkg/infrastructure/datastore"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	app := &cli.App{
		Name:  "cloud adapter",
		Usage: "run cloud adaptor server",
		Flags: append([]cli.Flag{
			&cli.BoolFlag{
				Name:  "testMode",
				Value: false,
				Usage: "A trigger to enable test mode.",
			},
			&cli.StringFlag{
				Name:    "logLevel",
				Value:   "debug",
				Usage:   "The level of logger.",
				EnvVars: []string{"LOG_LEVEL"},
			},
			&cli.StringFlag{
				Name:    "nsqd-server",
				Aliases: []string{"nsqd"},
				Value:   "127.0.0.1:4150",
				Usage:   "nsqd server address",
			},
			&cli.StringFlag{
				Name:    "nsq-lookupd-server",
				Aliases: []string{"lookupd"},
				Value:   "127.0.0.1:4161",
				Usage:   "nsq lookupd server address",
			},
			&cli.StringFlag{
				Name:    "listen",
				Aliases: []string{"l"},
				Value:   "127.0.0.1:8080",
				Usage:   "daemon server listen address",
				EnvVars: []string{"LISTEN"},
			},
		}, dbInfoFlag...),
		Action: run,
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Errorf("run cloud-adapter: %+v", err)
		os.Exit(1)
	}
}

func run(c *cli.Context) error {
	_, cancel := context.WithCancel(c.Context)
	defer cancel()

	config.Parse(c)
	config.SetLogLevel()

	db := datastore.NewDB()
	defer func() { _ = db.Close() }()
	datastore.AutoMigrate(db)

	//createChan := make(chan task.KubernetesConfigMessage, 10)
	//initChan := make(chan task.InitRainbondConfigMessage, 10)
	//updateChan := make(chan task.UpdateKubernetesConfigMessage, 10)
	//var taskProducer producer.TaskProducer
	//if os.Getenv("QUEUE_TYPE") == "nsq" {
	//	taskProducer = producer.NewTaskProducer()
	//	taskProducer.Start()
	//} else {
	//	taskProducer = producer.NewTaskChannelProducer(createChan, initChan, updateChan)
	//}

	//clusterUsecase := clustersvc.NewClusterUsecase(taskProducer)
	//createKubernetesTaskHandler := NewCreateKubernetesTaskHandler(clusterUsecase)
	//cloudInitTaskHandler := NewCloudInitTaskHandler(clusterUsecase)
	//cloudUpdateTaskHandler := NewCloudUpdateTaskHandler(clusterUsecase)
	//graph, err := initObj(ctx, db, taskProducer, clusterUsecase)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//if os.Getenv("QUEUE_TYPE") == "nsq" {
	//	msgConsumer := nsqc.NewTaskConsumer(config.C, createKubernetesTaskHandler, cloudInitTaskHandler)
	//	go msgConsumer.Start()
	//} else {
	//	msgConsumer := nsqc.NewTaskChannelConsumer(ctx, createChan, initChan, updateChan,
	//		createKubernetesTaskHandler, cloudInitTaskHandler, cloudUpdateTaskHandler)
	//	go msgConsumer.Start()
	//}

	r, err := initGin(db, nil)
	if err != nil {
		return err
	}
	logrus.Infof("start listen %s", c.String("listen"))
	go r.Run(c.String("listen"))

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	select {
	case <-term:
		logrus.Warn("Received SIGTERM, exiting gracefully...")
	}
	logrus.Info("See you next time!")
	return nil
}

func newGinEngine(router *handler.Router) *gin.Engine {
	r := router.NewRouter()
	r.Use(gin.Recovery())
	return r
}
