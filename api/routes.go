package api

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/priyankishorems/bollytics-go/api/handlers"
)

func SetupRoutes(h *handlers.Handlers) *echo.Echo {
	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(IPRateLimit(h))
	e.Use(ManageSession(h))
	e.Use(middleware.RemoveTrailingSlash())

	// limiterStore := middleware.NewRateLimiterMemoryStore(20)
	// e.Use(middleware.RateLimiter(limiterStore))

	e.HideBanner = true
	e.GET("/", h.HomeFunc)
	e.GET("/login", h.LoginHandler)
	e.GET("/callback", h.CallbackHandler)

	api := e.Group("/api")
	{
		tmdb := api.Group("/tmdb")
		{
			tmdb.GET("/actors/:name", h.SearchActorsHandler)
			tmdb.GET("/movies/:name", h.SearchMoviesHandler)
		}

		reddit := api.Group("/reddit")
		{
			reddit.GET("/temp", h.GetFromReddit)
			reddit.GET("/:sub/trending", h.GetTrendingWordsHandler)
			reddit.GET("/:sub/frequency", h.GetPostFrequencyHandler)
			reddit.GET("/:sub/:category/users", h.GetTopUsersHandler)
			reddit.GET("/:sub/:category/posts", h.GetTopPostsHandler)
		}

		scheduler, err := gocron.NewScheduler()
		if err != nil {
			log.Fatal("Error creating scheduler", err)
		}
		atTime := gocron.NewAtTime(23, 45, 0)
		atTimes := gocron.NewAtTimes(atTime)

		updateRedditPostsJob, err := scheduler.NewJob(gocron.DailyJob(1, atTimes), gocron.NewTask(func() {
			log.Info("Running updateRedditPostsJob")

			if err := h.UpdatePostsFromReddit(); err != nil {
				log.Error("Error updating posts from Reddit: ", err)
			}

			log.Info("updateRedditPostsJob completed")
		}))

		if err != nil {
			log.Fatal("Error creating job: ", err)
		}

		log.Info("updateRedditPostsJob started: ", updateRedditPostsJob.ID())

		scheduler.Start()

	}

	return e
}
