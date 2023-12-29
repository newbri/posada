package token

import (
	"github.com/newbri/posadamissportia/db"
	"time"
)

type Maker interface {
	CreateToken(username string, role *db.Role, duration time.Duration) (string, *Payload, error)
	VerifyToken(token string) (*Payload, error)
}
