package main

import (
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/newbri/posadamissportia/api"
	"github.com/newbri/posadamissportia/db"
	"github.com/newbri/posadamissportia/db/util"
	"github.com/newbri/posadamissportia/token"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	appConfig, err := util.NewYAMLConfiguration("app.yaml")
	if err != nil {
		log.Fatal().Msg("cannot create app configuration")
	}

	config, err := appConfig.GetConfig("dev")
	if err != nil {
		log.Fatal().Msg("cannot get app configuration")
	}

	if config.Name == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Msg("could not connect to the database")
	}

	runDBMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(conn)

	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		log.Fatal().Msg("cannot create token paseto")
	}

	runGinServer(store, tokenMaker, config)
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Msg("cannot create new migrate instance")
	}

	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal().Msg("failed to run migration up")
	}

	log.Info().Msg("db migration successfully")
}

func runGinServer(store db.Store, maker token.Maker, config *util.Config) {
	server := api.NewServer(store, maker, config)

	err := server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot start server")
	}
}
