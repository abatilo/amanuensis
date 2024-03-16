package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/abatilo/amanuensis/cmd/feedreader"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logLevel := new(slog.LevelVar)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	viper.SetEnvPrefix("AM")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	rootCmd := &cobra.Command{
		Use: "amanuesis",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if viper.GetBool(FlagVerbose) {
				logLevel.Set(slog.LevelDebug)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	rootCmd.PersistentFlags().BoolP(FlagVerbose, "v", false, "Enable global verbose logging")
	_ = viper.BindPFlags(rootCmd.PersistentFlags())

	rootCmd.AddCommand(feedreader.Cmd(ctx, logger))
	_ = rootCmd.Execute()
}
