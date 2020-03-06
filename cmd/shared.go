package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/scalify/puppet-master-gateway/pkg/database"
	"github.com/sirupsen/logrus"
	"github.com/rhinoman/couchdb-go"
	"github.com/streadway/amqp"
)

func setupLogger(logger *logrus.Logger, verbose bool) {
	if verbose {
		logger.SetLevel(logrus.DebugLevel)
		logger.Debug("Starting in debug level")
	} else {
		logger.SetLevel(logrus.InfoLevel)
		logger.Formatter = new(logrus.JSONFormatter)
	}
}

func connectQueue(logger *logrus.Logger, user, password, host string, port int) (*amqp.Connection, *amqp.Channel) {
	queueURI := fmt.Sprintf("amqp://%s:%s@%s:%d", user, password, host, port)
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

func connectJobDB(logger *logrus.Logger, cfg env) *database.JobDB {
	couch, err := couchdb.NewConnection(cfg.CouchDbHost, cfg.CouchDbPort, 1*time.Second)
	if err != nil {
		logger.Fatalf("Failed to open couchdb connection: %v", err)
	}

	for _, db := range []string{"_global_changes", "_metadata", "_replicator", "_users", "jobs"} {
		if err := couch.CreateDB(db, &couchdb.BasicAuth{Username: cfg.CouchDbUsername, Password: cfg.CouchDbPassword}); err != nil {
			if cErr, ok := err.(*couchdb.Error); ok {
				if cErr.StatusCode == 412 {
					logger.Debugf("Database %s already exists", db)
					continue
				}
			}
			logger.Fatalf("Failed to create database %s: %v", db, err)
		}
	}

	db := database.NewJobDB(couch.SelectDB("jobs", &couchdb.BasicAuth{Username: cfg.CouchDbUsername, Password: cfg.CouchDbPassword}))
	logger.Infof("Using database on http://%s:%d", cfg.CouchDbHost, cfg.CouchDbPort)

	return db
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
