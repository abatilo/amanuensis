package feedreader

import (
	"log/slog"

	"gorm.io/gorm"
)

const (
	FlagDBHost     = "db-host"
	FlagDBUser     = "db-user"
	FlagDBPassword = "db-password"
	FlagDBName     = "db-name"
)

type Config struct {
	logger *slog.Logger
	db     *gorm.DB
}
