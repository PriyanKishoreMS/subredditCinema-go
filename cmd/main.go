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
	"github.com/priyankishorems/bollytics-go/utils"
)

var validate validator.Validate

func main() {
	cfg := &utils.Config{}

	flag.IntVar(&cfg.Port, "port", 3000, "Server port")
	flag.StringVar(&cfg.Env, "env", "development", "Server port")

	flag.IntVar(&cfg.RateLimiter.Rps, "limiter-rps", 10, "Rate limiter max requests per second")
	flag.IntVar(&cfg.RateLimiter.Burst, "limiter-burst", 10, "Rate limiter max burst")
	flag.BoolVar(&cfg.RateLimiter.Enabled, "limiter-enabled", true, "Rate limiter enabled")

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
		Config:   *cfg,
		Validate: validate,
		Utils:    utils.NewUtils(),
		Data:     data.NewModel(db),
	}

	e := api.SetupRoutes(h)
	e.Server.ReadTimeout = time.Second * 10
	e.Server.WriteTimeout = time.Second * 20
	e.Server.IdleTimeout = time.Minute
	e.HideBanner = true
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", cfg.Port)))
}
