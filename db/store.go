package db

import "github.com/jackc/pgx/v5"

type Store interface {
	Querier
}

type SQLStore struct {
	*Queries
}

func NewStore(conn *pgx.Conn) Store {
	return &SQLStore{Queries: New(conn)}
}
