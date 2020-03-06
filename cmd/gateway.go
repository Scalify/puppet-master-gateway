package cmd

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/scalify/puppet-master-gateway/pkg/gateway"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type env struct {
	ListenPort      int    `default:"3000" split_words:"true"`
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

// gatewayCmd represents the gateway command
var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()
		ctx := newExitHandlerContext(logger)

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

		if err := server.Start(ctx, cfg.ListenPort); err != nil {
			logger.Fatalf("Failed to start gateway: %v", err)
		}

		<-ctx.Done()
	},
}

func init() {
	RootCmd.AddCommand(gatewayCmd)
}
