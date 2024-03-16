package feedreader

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Cmd(ctx context.Context, logger *slog.Logger) *cobra.Command {
	logger = logger.With("cmd", "feedreader")

	cmd := &cobra.Command{
		Use:   "feedreader",
		Short: "Component that reads RSS feeds",
		Run: func(cmd *cobra.Command, args []string) {
			dsn := fmt.Sprintf(
				"host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
				viper.GetString(FlagDBHost),
				viper.GetString(FlagDBUser),
				viper.GetString(FlagDBPassword),
				viper.GetString(FlagDBName),
			)
			db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
			if err != nil {
				logger.Error("failed to connect to database", "error", err)
				return
			}

			server := NewServer(ctx, Config{
				logger: logger,
				db:     db,
			})
			if err := server.Start(); err != http.ErrServerClosed {
				logger.Error("failed to start server", "error", err)
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().String(FlagDBHost, "localhost", "Postgres host to connect to")
	cmd.PersistentFlags().String(FlagDBUser, "postgres", "Postgres user to connect as")
	cmd.PersistentFlags().String(FlagDBPassword, "local_password", "Password for the Postgres user")
	cmd.PersistentFlags().String(FlagDBName, "postgres", "Name of the database to connect to")
	_ = viper.BindPFlags(cmd.PersistentFlags())
	return cmd
}
