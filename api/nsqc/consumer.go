package nsqc

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/nsqio/go-nsq"
	"github.com/prometheus/common/log"
	"goodrain.com/cloud-adaptor/api/config"
	"goodrain.com/cloud-adaptor/api/handler"
	"goodrain.com/cloud-adaptor/util/constants"
)

//TaskConsumer task producer
type TaskConsumer interface {
	Start() error
}

// taskConsumer -
type taskConsumer struct {
	config                      *config.Config
	createKubernetesTaskHandler handler.CreateKubernetesTaskHandler
	cloudInitTaskHandler        handler.CloudInitTaskHandler
}

// NewTaskConsumer creates a new consumer.
func NewTaskConsumer(config *config.Config, createHandler handler.CreateKubernetesTaskHandler, initHandler handler.CloudInitTaskHandler) TaskConsumer {
	return &taskConsumer{
		config:                      config,
		createKubernetesTaskHandler: createHandler,
		cloudInitTaskHandler:        initHandler,
	}
}

func (c *taskConsumer) Start() error {
	config := nsq.NewConfig()
	//cloud init handler
	initConsumer, err := nsq.NewConsumer(constants.CloudInit, "default", config)
	if err != nil {
		return err
	}

	// Set the Handler for messages received by this Consumer. Can be called multiple times.
	// See also AddConcurrentHandlers.
	initConsumer.AddHandler(c.cloudInitTaskHandler)

	// Use nsqlookupd to discover nsqd instances.
	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
	err = initConsumer.ConnectToNSQLookupd(c.config.NSQConfig.NsqLookupdAddress)
	if err != nil {
		return err
	}
	defer initConsumer.Stop()

	createConsumer, err := nsq.NewConsumer(constants.CloudCreate, "default", config)
	if err != nil {
		return err
	}
	// Set the Handler for messages received by this Consumer. Can be called multiple times.
	// See also AddConcurrentHandlers.
	createConsumer.AddHandler(c.createKubernetesTaskHandler)

	// Use nsqlookupd to discover nsqd instances.
	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
	err = createConsumer.ConnectToNSQLookupd(c.config.NSQConfig.NsqLookupdAddress)
	if err != nil {
		return err
	}

	// Gracefully stop the consumer.
	defer createConsumer.Stop()
	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	select {
	case v := <-initConsumer.StopChan:
		log.Errorf("Received initConsumer stop signal %d, exiting gracefully...", v)
	case v := <-createConsumer.StopChan:
		log.Errorf("Received createConsumer stop signal %d, exiting gracefully...", v)
	case <-term:
		log.Warn("Received SIGTERM, exiting gracefully...")
	}
	log.Info("See you next time!")
	return nil
}
