package data

import "github.com/jackc/pgx/v5"

type TaskModel struct {
	DB *pgx.Conn
}
