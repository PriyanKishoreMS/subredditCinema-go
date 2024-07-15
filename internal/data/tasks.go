package data

import pgx "github.com/jackc/pgx/v5"

type TaskModel struct {
	DB *pgx.Conn
}

func (t TaskModel) CreateTask() error {
	return nil
}
