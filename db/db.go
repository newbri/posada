package db

import (
	"github.com/jackc/pgx/v5"
)

type Queries struct {
	conn *pgx.Conn
}

func New(conn *pgx.Conn) *Queries {
	return &Queries{conn: conn}
}
