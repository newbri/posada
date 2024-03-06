package main

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newbri/posadamissportia/db"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/newbri/posadamissportia/api"
	"github.com/newbri/posadamissportia/configuration"
	"github.com/newbri/posadamissportia/token"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	env := os.Getenv("POSADA_ENV")
	yamlConfig := configuration.NewYAMLConfiguration("app.yaml", env)

	if yamlConfig.GetConfig().Name == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	pool, err := pgxpool.New(context.Background(), yamlConfig.GetConfig().DBSource)
	if err != nil {
		log.Fatal().Msg("could not connect to the database")
	}
	defer pool.Close()

	runDBMigration(yamlConfig.GetConfig().MigrationURL, yamlConfig.GetConfig().DBSource)

	store := db.NewStore(pool)

	tokenMaker, err := token.NewPasetoMaker(yamlConfig.GetConfig().TokenSymmetricKey)
	if err != nil {
		log.Fatal().Msg("cannot create token paseto")
	}

	server := api.NewServer(store, tokenMaker, yamlConfig)
	err = server.Start(yamlConfig.GetConfig().HTTPServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot start server")
	}
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Msg("cannot create new migrate instances")
	}

	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal().Msg("failed to run migration up")
	}

	log.Info().Msg("db migration successfully")
}
