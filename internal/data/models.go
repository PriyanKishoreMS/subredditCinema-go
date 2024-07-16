package data

import (
	"context"
	"time"

	pgx "github.com/jackc/pgx/v5"
)

func Handlectx() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	return ctx, cancel
}

type Models struct {
	Posts PostModel
}

func NewModel(db *pgx.Conn) Models {
	return Models{
		Posts: PostModel{DB: db},
	}
}
