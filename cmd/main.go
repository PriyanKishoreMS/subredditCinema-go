package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/gommon/log"
	"github.com/priyankishorems/bollytics-go/api"
	"github.com/priyankishorems/bollytics-go/api/handlers"
	"github.com/priyankishorems/bollytics-go/internal/data"
	"github.com/priyankishorems/bollytics-go/utils"
	sw "github.com/toadharvard/stopwords-iso"
	"github.com/vartanbeno/go-reddit/v2/reddit"
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

	db := data.PSQLDB{}
	dbPool, err := db.Open()
	if err != nil {
		log.Fatalf("error in opening db; %v", err)
	}
	defer dbPool.Close()

	validate = *validator.New()

	tmdbClient, err := tmdb.Init(utils.TMDBKey)
	if err != nil {
		log.Fatalf("error in initializing tmdb client; %v", err)
	}
	log.Info("TMDB client initialized")

	tmdbClient.SetClientConfig(*utils.HttpClientConfig)
	tmdbClient.SetClientAutoRetry()

	// redditBot, err := graw.NewBotFromAgentFile("graw.ini", 0)
	// if err != nil {
	// 	log.Fatalf("error in initializing reddit bot; %v", err)
	// }

	// log.Info("Graw Bot initialized")

	redditCredentials := reddit.Credentials{
		ID:       utils.RedditId,
		Secret:   utils.RedditSecret,
		Username: utils.RedditUsername,
		Password: utils.RedditPassword,
	}

	redditClient, err := reddit.NewClient(redditCredentials)
	if err != nil {
		log.Fatalf("error in initializing go-reddit client; %v", err)
	}

	stopword, err := sw.NewStopwordsMapping()
	if err != nil {
		log.Fatalf("error in initializing stopwords; %v", err)
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 6 * time.Hour
	sessionManager.IdleTimeout = 20 * time.Minute
	sessionManager.Cookie.Name = "login-session"
	// sessionManager.Cookie.Domain = "localhost:3000"
	sessionManager.Cookie.HttpOnly = true
	// sessionManager.Cookie.Path = "/example/"
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = http.SameSiteStrictMode
	sessionManager.Cookie.Secure = true

	log.Info("Reddit client initialized")

	h := &handlers.Handlers{
		Config:   *cfg,
		Validate: validate,
		Utils:    utils.NewUtils(),
		Data:     data.NewModel(dbPool),
		Tmdb:     tmdbClient,
		// RedditBot:      redditBot,
		Reddit:         redditClient,
		Stopword:       stopword,
		SessionManager: sessionManager,
	}

	e := api.SetupRoutes(h)
	e.Server.ReadTimeout = time.Second * 10
	e.Server.WriteTimeout = time.Second * 20
	e.Server.IdleTimeout = time.Minute
	e.HideBanner = true
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", cfg.Port)))
}
