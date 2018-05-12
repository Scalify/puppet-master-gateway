package cmd

import (
	"github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"gitlab.com/scalifyme/puppet-master/puppet-master/pkg/gateway"
)

type gatewayEnv struct {
	CouchEnv
	SharedEnv
	ListenPort int `default:"3000" split_words:"true"`
}

// gatewayCmd represents the gateway command
var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()
		ctx := newExitHandlerContext(logger)

		var cfg gatewayEnv
		if err := envconfig.Process("", &cfg); err != nil {
			logger.Fatal(err)
		}

		setupLogger(logger, cfg.SharedEnv)
		db := connectJobDB(logger, cfg.CouchEnv)

		gateway, err := gateway.NewServer(db, logger.WithFields(logrus.Fields{}))
		if err != nil {
			logger.Fatalf("Failed to create gateway: %v", err)
		}

		logger.Infof("Listening on port %d", cfg.ListenPort)

		if err := gateway.Start(ctx, cfg.ListenPort); err != nil {
			logger.Fatalf("Failed to start gateway: %v", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(gatewayCmd)
}
