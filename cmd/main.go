package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/gommon/log"
	"github.com/priyankishorems/bollytics-go/api"
	"github.com/priyankishorems/bollytics-go/api/handlers"
	"github.com/priyankishorems/bollytics-go/internal/data"
)

type config struct {
	port int
	env  string
}

var validate validator.Validate

func main() {
	cfg := &config{}

	flag.IntVar(&cfg.port, "port", 3000, "Server port")
	flag.StringVar(&cfg.env, "env", "development", "Server port")

	flag.Parse()
	log.SetHeader("${time_rfc3339} ${level}")

	conn := data.PSQLDB{}
	db, err := conn.Open()
	if err != nil {
		log.Fatalf("error in opening db; %v", err)
	}
	defer db.Close(context.Background())

	validate = *validator.New()

	h := &handlers.Handlers{
		Validate: validate,
	}

	e := api.SetupRoutes(h)
	e.Server.ReadTimeout = time.Second * 10
	e.Server.WriteTimeout = time.Second * 20
	e.Server.IdleTimeout = time.Minute
	e.HideBanner = true
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", cfg.port)))
}
