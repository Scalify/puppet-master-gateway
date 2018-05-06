package cmd

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
	"github.com/rhinoman/couchdb-go"
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
	"gitlab.com/scalifyme/puppet-master/gateway/pkg/gateway/db"
	"gitlab.com/scalifyme/puppet-master/gateway/pkg/gateway/server"
)

type serverEnv struct {
	CouchDbHost     string `required:"true" split_words:"true"`
	CouchDbPort     int    `required:"true" split_words:"true"`
	CouchDbUsername string `required:"true" split_words:"true"`
	CouchDbPassword string `required:"true" split_words:"true"`
	QueueHost       string `required:"true" split_words:"true"`
	QueuePort       int    `required:"true" split_words:"true"`
	QueueUsername   string `required:"true" split_words:"true"`
	QueuePassword   string `required:"true" split_words:"true"`
	ListenPort      int    `default:"3000" split_words:"true"`
	Verbose         bool   `default:"false" split_words:"true" def`
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()

		var cfg serverEnv
		if err := envconfig.Process("", &cfg); err != nil {
			logger.Fatal(err)
		}

		if cfg.Verbose {
			logger.SetLevel(logrus.DebugLevel)
			logger.Debug("Starting in debug level")
		} else {
			logger.SetLevel(logrus.InfoLevel)
		}

		couch, err := couchdb.NewConnection(cfg.CouchDbHost, cfg.CouchDbPort, time.Duration(1*time.Second))
		if err != nil {
			logger.Fatalf("Failed to open couchdb connection: %v", err)
		}
		db := db.NewJobDB(couch.SelectDB("jobs", &couchdb.BasicAuth{Username: cfg.CouchDbUsername, Password: cfg.CouchDbPassword}))

		queueURI := fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.QueueUsername, cfg.QueuePassword, cfg.QueueHost, cfg.QueuePort)
		logger.Infof("Using Queue on %q", queueURI)
		queueConn, err := amqp.Dial(queueURI)
		if err != nil {
			logger.Fatalf("Failed to connect to queue: %v", err)
		}
		defer queueConn.Close()

		queueChannel, err := queueConn.Channel()
		if err != nil {
			logger.Fatalf("Failed to open channel on queue connection: %v", err)
		}

		server, err := server.NewServer(db, queueChannel, logger.WithFields(logrus.Fields{}))
		if err != nil {
			logger.Fatalf("Failed to create server: %v", err)
		}

		logger.Infof("Listening on port %d", cfg.ListenPort)
		server.Serve(cfg.ListenPort)
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)
}
