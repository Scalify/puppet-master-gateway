package cmd

import (
	"github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"gitlab.com/scalifyme/puppet-master/puppet-master/pkg/coordinator"
)

type coordinatorEnv struct {
	CouchEnv
	QueueEnv
	SharedEnv
}

// coordinatorCmd represents the coordinator command
var coordinatorCmd = &cobra.Command{
	Use:   "coordinator",
	Short: "Handles queue messages to control the executor and consume the results",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()
		ctx := newExitHandlerContext(logger)

		var cfg coordinatorEnv
		if err := envconfig.Process("", &cfg); err != nil {
			logger.Fatal(err)
		}

		setupLogger(logger, cfg.SharedEnv)
		db := connectJobDB(logger, cfg.CouchEnv)
		conn, queue := connectQueue(logger, cfg.QueueEnv)
		defer func() {
			if err := conn.Close(); err != nil {
				logger.Fatalf("Failed to close queue connection: %v", err)
			}
		}()

		server, err := coordinator.New(logger.WithFields(logrus.Fields{}), db, queue)
		if err != nil {
			logger.Fatalf("Failed to create server: %v", err)
		}

		if err := server.Start(ctx); err != nil {
			logger.Fatalf("Failed to start coordinator: %v", err)
		}

	},
}

func init() {
	RootCmd.AddCommand(coordinatorCmd)
}
