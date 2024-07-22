package data

import (
	"context"
	"time"

	pgx "github.com/jackc/pgx/v5/pgxpool"
)

func Handlectx() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	return ctx, cancel
}

type Models struct {
	Posts   PostModel
	Users   UserModel
	Polls   PollsModel
	Surveys SurveysModel
}

func NewModel(db *pgx.Pool) Models {
	return Models{
		Posts:   PostModel{DB: db},
		Users:   UserModel{DB: db},
		Polls:   PollsModel{DB: db},
		Surveys: SurveysModel{DB: db},
	}
}
