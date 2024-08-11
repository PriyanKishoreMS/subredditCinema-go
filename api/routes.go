package api

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/priyankishorems/bollytics-go/api/handlers"
	"github.com/priyankishorems/bollytics-go/jobs"
)

func SetupRoutes(h *handlers.Handlers) *echo.Echo {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowCredentials: true,
	}))
	e.Use(IPRateLimit(h))
	e.Use(middleware.RemoveTrailingSlash())
	// e.Use(CacheControlWordCloud())
	e.Static("/public", "public")

	e.HideBanner = true
	e.GET("/", h.HomeFunc)
	e.GET("/login", h.LoginHandler)
	e.GET("/callback", h.CallbackHandler)
	e.GET("/refresh", h.RefreshTokenHandler)

	api := e.Group("/api")
	{
		survey := api.Group("/survey")
		{
			survey.POST("/create", h.CreateSurveyHandler)
			survey.POST("/response/:survey_id", h.CreateSurveyResponsesHandler, Authenticate(*h))
			survey.GET("/:survey_id", h.GetSurveyByIDHandler)
			survey.GET("", h.GetAllSurveysHandler)
			survey.GET("/results/:survey_id", h.GetSurveyResultsHandler)
		}

		poll := api.Group("/poll")
		{
			poll.GET("/:sub/all", h.GetAllPollsHandler)
			poll.GET("/:poll_id", h.GetPollByIDHandler)
			// todo Should Add Middleware to Authenticate User before deploying
			poll.POST("/create", h.CreatePollHandler)
			poll.POST("/vote/:poll_id/:option_id", h.CreatePollVoteHandler)
			poll.DELETE("/delete/:poll_id", h.DeletePollByCreatorHandler)
		}

		tmdb := api.Group("/tmdb")
		{
			tmdb.GET("/actors/:name", h.SearchActorsHandler)
			tmdb.GET("/movies/:name", h.SearchMoviesHandler)
		}

		reddit := api.Group("/reddit")
		{
			reddit.GET("/temp", h.GetFromReddit)
			reddit.GET("/:sub/trending", h.GetTrendingWordsHandlerWeb)
			reddit.GET("/:sub/frequency", h.GetPostFrequencyHandler)
			reddit.GET("/:sub/:category/users", h.GetTopUsersHandler)
			reddit.GET("/:sub/:category/posts", h.GetTopPostsHandler)
		}

		scheduler, err := gocron.NewScheduler()
		if err != nil {
			log.Fatal("Error creating scheduler", err)
		}
		updatePostsAtTime := gocron.NewAtTime(23, 40, 00)
		updatePostsAtTimes := gocron.NewAtTimes(updatePostsAtTime)
		updateWordCloudAtTime := gocron.NewAtTime(23, 40, 30)
		updateWordCloudAtTimes := gocron.NewAtTimes(updateWordCloudAtTime)

		updateRedditPostsJob, err := jobs.UpdateRedditPostsJob(*h, scheduler, updatePostsAtTimes)
		if err != nil {
			log.Fatal("Error creating job: ", err)
		}

		updateWordCloudsJob, err := jobs.UpdateWordClouds(*h, scheduler, updateWordCloudAtTimes)
		if err != nil {
			log.Fatal("Error creating job: ", err)
		}

		log.Info("updateRedditPostsJob started: ", updateRedditPostsJob.ID())
		log.Info("updateWordCloudsJob started: ", updateWordCloudsJob.ID())

		scheduler.Start()

	}

	return e
}
