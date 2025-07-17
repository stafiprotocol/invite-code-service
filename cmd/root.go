package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

var (
	appName = "invite-code-service"
)

const (
	flagLogLevel   = "log_level"
	flagConfigPath = "config"

	defaultConfigPath = "./config.toml"
)

// NewRootCmd returns the root command.
func NewRootCmd() *cobra.Command {
	// RootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   appName,
		Short: "invite code services",
	}

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, segments []string) error {
		return nil
	}

	rootCmd.AddCommand(
		startApiCmd(),
		bindCmd(),
	)

	return rootCmd
}

func Execute() {

	rootCmd := NewRootCmd()
	rootCmd.SilenceUsage = true
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	ctx := context.Background()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
