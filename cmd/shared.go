package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/rhinoman/couchdb-go"
	"github.com/streadway/amqp"
	"gitlab.com/scalifyme/puppet-master/puppet-master/pkg/database"
)

func setupLogger(logger *logrus.Logger, cfg SharedEnv) {
	if cfg.Verbose {
		logger.SetLevel(logrus.DebugLevel)
		logger.Debug("Starting in debug level")
	} else {
		logger.SetLevel(logrus.InfoLevel)
		logger.Formatter = new(logrus.JSONFormatter)
	}
}

func connectJobDB(logger *logrus.Logger, cfg CouchEnv) *database.JobDB {
	couch, err := couchdb.NewConnection(cfg.CouchDbHost, cfg.CouchDbPort, 1*time.Second)
	if err != nil {
		logger.Fatalf("Failed to open couchdb connection: %v", err)
	}

	for _, db := range []string{"_global_changes", "_metadata", "_replicator", "_users", "jobs"} {
		if err := couch.CreateDB(db, &couchdb.BasicAuth{Username: cfg.CouchDbUsername, Password: cfg.CouchDbPassword}); err != nil {
			logger.Errorf("Failed to create database %s: %v", db, err)
		}
	}

	db := database.NewJobDB(couch.SelectDB("jobs", &couchdb.BasicAuth{Username: cfg.CouchDbUsername, Password: cfg.CouchDbPassword}))
	logger.Infof("Using database on http://%s:%d", cfg.CouchDbHost, cfg.CouchDbPort)

	return db
}

func connectQueue(logger *logrus.Logger, cfg QueueEnv) (*amqp.Connection, *amqp.Channel) {
	queueURI := fmt.Sprintf("amqp://%s:%s@%s:%d", cfg.QueueUsername, cfg.QueuePassword, cfg.QueueHost, cfg.QueuePort)
	logger.Infof("Using Queue on %s", queueURI)

	queueConn, err := amqp.Dial(queueURI)
	if err != nil {
		logger.Fatalf("Failed to connect to queue: %v", err)
	}

	queueChannel, err := queueConn.Channel()
	if err != nil {
		logger.Fatalf("Failed to open channel on queue connection: %v", err)
	}

	return queueConn, queueChannel
}

func newExitHandlerContext(logger *logrus.Logger) context.Context {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-c
		defer cancel()
		logger.Info("shutting down")
	}()

	return ctx
}
