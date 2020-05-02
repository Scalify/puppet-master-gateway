package main

import (
	"fmt"
	"time"

	"github.com/aklinkert/go-exitcontext"
	"github.com/kelseyhightower/envconfig"
	"github.com/rhinoman/couchdb-go"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"github.com/scalify/puppet-master-gateway/pkg/database"
	"github.com/scalify/puppet-master-gateway/pkg/gateway"
)

type env struct {
	ListenPort      uint   `default:"3000" split_words:"true"`
	Verbose         bool   `default:"false" split_words:"true"`
	EnableAPI       bool   `default:"true" split_words:"true" envconfig:"ENABLE_API"`
	EnableJobs      bool   `default:"true" split_words:"true"`
	QueueHost       string `required:"true" split_words:"true"`
	QueuePort       int    `required:"true" split_words:"true"`
	QueueUsername   string `required:"true" split_words:"true"`
	QueuePassword   string `required:"true" split_words:"true"`
	CouchDbHost     string `required:"true" split_words:"true"`
	CouchDbPort     int    `required:"true" split_words:"true"`
	CouchDbUsername string `required:"true" split_words:"true"`
	CouchDbPassword string `required:"true" split_words:"true"`
	APIToken        string `required:"true" split_words:"true" envconfig:"API_TOKEN"`
}

func main() {
	logger := logrus.New()
	ctx := exitcontext.New()

	var cfg env
	if err := envconfig.Process("", &cfg); err != nil {
		logger.Fatal(err)
	}

	if !cfg.EnableAPI && !cfg.EnableJobs {
		logger.Fatal("Either API or background job processing needs to be enabled")
	}

	conn, queue := connectQueue(logger, cfg.QueueUsername, cfg.QueuePassword, cfg.QueueHost, cfg.QueuePort)
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Fatalf("Failed to close queue connection: %v", err)
		}
	}()

	setupLogger(logger, cfg.Verbose)
	db := connectJobDB(logger, cfg)

	server, err := gateway.NewServer(db, queue, logger.WithFields(logrus.Fields{}), cfg.APIToken, cfg.EnableAPI, cfg.EnableJobs)
	if err != nil {
		logger.Fatalf("Failed to create gateway: %v", err)
	}

	logger.Infof("Listening on port %v", cfg.ListenPort)

	if err := server.Start(ctx, cfg.ListenPort); err != nil {
		logger.Fatalf("Failed to start gateway: %v", err)
	}

	<-ctx.Done()
}

func setupLogger(logger *logrus.Logger, verbose bool) {
	logger.Formatter = new(logrus.JSONFormatter)

	if verbose {
		logger.SetLevel(logrus.DebugLevel)
		logger.Debug("Starting in debug level")
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
