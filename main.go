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
	yamlConfig := util.NewYAMLConfiguration("app.yaml", "dev")

	if yamlConfig.GetConfig().Name == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(yamlConfig.GetConfig().DBDriver, yamlConfig.GetConfig().DBSource)
	if err != nil {
		log.Fatal().Msg("could not connect to the database")
	}

	runDBMigration(yamlConfig.GetConfig().MigrationURL, yamlConfig.GetConfig().DBSource)

	store := db.NewStore(conn)

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
		log.Fatal().Msg("cannot create new migrate instance")
	}

	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal().Msg("failed to run migration up")
	}

	log.Info().Msg("db migration successfully")
}
