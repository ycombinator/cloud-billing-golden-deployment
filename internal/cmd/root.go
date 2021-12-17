package cmd

import (
	"fmt"

	"go.uber.org/zap/zapcore"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/logging"

	"go.uber.org/zap"

	"github.com/spf13/cobra"
)

const flagLogLevel = "log-level"

var rootCmd = &cobra.Command{
	Use:   "ecbgd",
	Short: "ecbgd is the Elastic Cloud Billing Golden Deployment CLI",
	Long: "The Elastic Cloud Billing Golden Deployment CLI manages golden " +
		"deployments for validating metering and billing implementations.",
}

func init() {
	cobra.OnInitialize(initLogging)
	rootCmd.PersistentFlags().StringP(flagLogLevel, "l", "info", "log level")
	rootCmd.AddCommand(serverCmd)
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("could not execute root command: %w", err)
	}

	return nil

}
func initLogging() {
	logLevel, _ := rootCmd.PersistentFlags().GetString(flagLogLevel)

	var level zap.AtomicLevel
	level.UnmarshalText([]byte(logLevel))

	cfg := zap.Config{
		Level:            level,
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "@timestamp",
			NameKey:        "name",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			LineEnding:     "\n",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	logging.Logger = logger
}
