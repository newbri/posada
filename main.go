package main

import (
	"database/sql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/newbri/posadamissportia/api"
	"github.com/newbri/posadamissportia/configuration"
	"github.com/newbri/posadamissportia/db"
	"github.com/newbri/posadamissportia/db/util"
	"github.com/newbri/posadamissportia/token"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Msg("Error loading .env file")
	}

	env := os.Getenv("POSADA_ENV")
	yamlConfig := configuration.NewYAMLConfiguration("app.yaml", env)

	if yamlConfig.GetConfig().Name == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(yamlConfig.GetConfig().DBDriver, yamlConfig.GetConfig().DBSource)
	if err != nil {
		log.Fatal().Msg("could not connect to the database")
	}

	util.RunDBMigration(yamlConfig.GetConfig().MigrationURL, yamlConfig.GetConfig().DBSource)

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
