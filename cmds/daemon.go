package cmds

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"

	"fmt"
	"log"

	"os"
	"reflect"

	"github.com/facebookgo/inject"
	"github.com/jinzhu/gorm"

	"goodrain.com/cloud-adaptor/api/config"
	"goodrain.com/cloud-adaptor/api/infrastructure/datastore"
	"goodrain.com/cloud-adaptor/api/nsqc"
	"goodrain.com/cloud-adaptor/api/nsqc/producer"
	"goodrain.com/cloud-adaptor/task"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	clusterrepo "goodrain.com/cloud-adaptor/api/cluster/repository"
	clustersvc "goodrain.com/cloud-adaptor/api/cluster/usecase"
	myrouter "goodrain.com/cloud-adaptor/api/infrastructure/router"
	openapihdl "goodrain.com/cloud-adaptor/api/openapi/handler"
)

var daemonCommand = &cli.Command{
	Name: "daemon",
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
		},
	}, dbInfoFlag...),
	Action: runDaemon,
	Usage:  "run cloud adaptor daemon server",
}

func runDaemon(ctx *cli.Context) error {
	config.Parse(ctx)
	config.SetLogLevel()
	db := datastore.NewDB()
	defer func() { _ = db.Close() }()
	datastore.AutoMigrate(db)

	daemonCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	createChan := make(chan task.KubernetesConfigMessage, 10)
	initChan := make(chan task.InitRainbondConfigMessage, 10)

	var taskProducer producer.TaskProducer
	if os.Getenv("QUEUE_TYPE") == "nsq" {
		taskProducer = producer.NewTaskProducer()
		taskProducer.Start()
	} else {
		taskProducer = producer.NewTaskChannelProducer(createChan, initChan)
	}

	clusterUsecase := clustersvc.NewClusterUsecase(taskProducer)
	createKubernetesTaskHandler := NewCreateKubernetesTaskHandler(clusterUsecase)
	cloudInitTaskHandler := NewCloudInitTaskHandler(clusterUsecase)
	graph, err := initObj(daemonCtx, db, taskProducer, clusterUsecase)
	if err != nil {
		log.Fatal(err)
	}

	if os.Getenv("QUEUE_TYPE") == "nsq" {
		msgConsumer := nsqc.NewTaskConsumer(config.C, createKubernetesTaskHandler, cloudInitTaskHandler)
		go msgConsumer.Start()
	} else {
		msgConsumer := nsqc.NewTaskChannelConsumer(daemonCtx, createChan, initChan, createKubernetesTaskHandler, cloudInitTaskHandler)
		go msgConsumer.Start()
	}

	mr, err := GetRouter(graph)
	if err != nil {
		log.Fatal(err)
	}

	r := mr.NewRouter()
	r.Use(gin.Recovery())
	go r.Run(ctx.String("listen"))

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	select {
	case <-term:
		logrus.Warn("Received SIGTERM, exiting gracefully...")
	}
	logrus.Info("See you next time!")
	return nil
}

// initiate objects and provide them to the Graph if necessary
func initObj(ctx context.Context, db *gorm.DB, taskProducer, clusterUsecase interface{}) (*inject.Graph, error) {
	// Typically an application will have exactly one object graph, and
	// you will create it and use it within a main function:
	var g inject.Graph

	if err := g.Provide(
		&inject.Object{Value: config.C},
		&inject.Object{Value: db},
		&inject.Object{Name: "router", Value: myrouter.New()},
		//producer
		&inject.Object{Name: "TaskProducer", Value: taskProducer},
		// handler
		&inject.Object{Value: clusterUsecase},
		// repository
		&inject.Object{Value: clusterrepo.NewCloudAccessKeyRepo(nil)},
		&inject.Object{Value: clusterrepo.NewCreateKubernetesTaskRepo(nil)},
		&inject.Object{Value: clusterrepo.NewInitRainbondRegionTaskRepo(nil)},
		&inject.Object{Value: clusterrepo.NewTaskEventRepo(nil)},
		// openapi
		&inject.Object{Value: openapihdl.NewClusterHandler()},
	); err != nil {
		return nil, fmt.Errorf("provide objects to the Graph: %v", err)
	}

	if err := g.Populate(); err != nil {
		return nil, fmt.Errorf("populate the incomplete Objects: %v", err)
	}
	return &g, nil
}

// GetRouter return my router.
func GetRouter(g *inject.Graph) (*myrouter.Router, error) {
	obj := GetObject(g, "router")
	if obj == nil {
		return nil, fmt.Errorf("my router not found")
	}
	r, ok := obj.Value.(*myrouter.Router)
	if !ok {
		return nil, fmt.Errorf("expetcd *router.Router, but got %v", reflect.TypeOf(obj.Value))
	}
	return r, nil
}

// GetObject returns an object in grach based on the given name.
func GetObject(g *inject.Graph, name string) *inject.Object {
	for _, obj := range g.Objects() {
		if obj.Name == name {
			return obj
		}
	}
	return nil
}
