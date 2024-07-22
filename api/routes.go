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

	e.HideBanner = true
	e.GET("/", h.HomeFunc)
	e.GET("/login", h.LoginHandler)
	e.GET("/callback", h.CallbackHandler)
	e.GET("/verify", h.VerifySession, AuthenticateUserSession(h))

	api := e.Group("/api")
	{
		survey := api.Group("/survey")
		{
			survey.GET("/:survey_id", h.GetSurveyByIDHandler)
			// todo Should Add Middleware to Authenticate User before deploying
			survey.GET(("/response/:survey_id"), h.GetSurveyResponsesByIDHandler)
			survey.POST("/create", h.CreateSurveyHandler)
			survey.POST("/questions/:survey_id", h.CreateSurveyQuestionsHandler)
			survey.POST("/response/:survey_id", h.CreateSurveyResponseHandler)
		}

		poll := api.Group("/poll")
		{
			poll.GET("/all", h.GetAllPollsHandler)
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
