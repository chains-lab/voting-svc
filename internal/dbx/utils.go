package dbx

import (
	"database/sql"
	"embed"

	"github.com/chains-lab/voting-svc/internal/config"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

type txKeyType struct{}

var TxKey = txKeyType{}

//go:embed migrations/*.sql
var Migrations embed.FS

var migrations = &migrate.EmbedFileSystemMigrationSource{
	FileSystem: Migrations,
	Root:       "migrations",
}

func MigrateUp(cfg config.Config) error {
	db, err := sql.Open("postgres", cfg.Database.SQL.URL)

	applied, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return errors.Wrap(err, "failed to apply migrations")
	}
	logrus.WithField("applied", applied).Info("migrations applied")
	return nil
}

func MigrateDown(cfg config.Config) error {
	db, err := sql.Open("postgres", cfg.Database.SQL.URL)

	applied, err := migrate.Exec(db, "postgres", migrations, migrate.Down)
	if err != nil {
		return errors.Wrap(err, "failed to apply migrations")
	}
	logrus.WithField("applied", applied).Info("migrations applied")
	return nil
}
