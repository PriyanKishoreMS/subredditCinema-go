package data

import (
	"context"
	"fmt"

	pgx "github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/gommon/log"
	"github.com/priyankishorems/bollytics-go/utils"
)

type PSQLDB struct {
	database string
	username string
	pwd      string
	port     string
	host     string
}

func (m PSQLDB) Open() (*pgx.Pool, error) {
	c := PSQLDB{
		database: utils.DBName,
		username: utils.DBUsername,
		pwd:      utils.DBPassword,
		port:     utils.DBPort,
		host:     utils.DBHost,
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", c.username, c.pwd, c.host, c.port, c.database)
	dbpool, err := pgx.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	var dbName string
	err = dbpool.QueryRow(context.Background(), "SELECT current_database()").Scan(&dbName)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Connected to database: %s", dbName)

	return dbpool, nil
}
