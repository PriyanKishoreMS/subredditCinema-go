package api

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/go-co-op/gocron/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/priyankishorems/bollytics-go/api/handlers"
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
	e.GET("/verify", h.VerifySession)
	e.GET("/refresh", h.RefreshTokenHandler)

	api := e.Group("/api")
	{
		survey := api.Group("/survey")
		{
			survey.POST("/create", h.CreateSurveyHandler)
			survey.POST("/response/:survey_id", h.CreateSurveyResponsesHandler)
			survey.GET("/:survey_id", h.GetSurveyByIDHandler)
			survey.GET("/all", h.GetAllSurveysHandler)
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
		updateWordCloudAtTime := gocron.NewAtTime(23, 42, 00)
		updateWordCloudAtTimes := gocron.NewAtTimes(updateWordCloudAtTime)

		updateRedditPostsJob, err := updateRedditPostsJob(*h, scheduler, updatePostsAtTimes)
		if err != nil {
			log.Fatal("Error creating job: ", err)
		}

		updateWordCloudsJob, err := updateWordClouds(*h, scheduler, updateWordCloudAtTimes)
		if err != nil {
			log.Fatal("Error creating job: ", err)
		}

		log.Info("updateRedditPostsJob started: ", updateRedditPostsJob.ID())
		log.Info("updateWordCloudsJob started: ", updateWordCloudsJob.ID())

		scheduler.Start()

	}

	return e
}

func updateWordClouds(h handlers.Handlers, scheduler gocron.Scheduler, atTimes gocron.AtTimes) (gocron.Job, error) {
	job, err := scheduler.NewJob(gocron.DailyJob(1, atTimes), gocron.NewTask(func() error {
		log.Info("Running updateWordClouds")

		subs := []string{"kollywood", "bollywood", "tollywood", "MalayalamMovies"}

		for _, sub := range subs {
			words, err := h.GetTrendingWordsHandler(sub, "month")
			if err != nil {
				log.Error("Error updating word clouds: ", err)
			}

			jsonWords := map[string][]handlers.WordCount{
				sub: words,
			}

			jsonBytes, err := json.Marshal(jsonWords)
			if err != nil {
				log.Error("Error marshalling json: ", err)
				return err
			}

			cmd := exec.Command("wordcloud/py-venv/bin/python", "wordcloud/main.py")
			cmd.Stdin = strings.NewReader(string(jsonBytes))

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err = cmd.Run()
			if err != nil {
				log.Error("Error running wordcloud script: ", err)
				log.Error("Python script stdout: ", stdout.String())
				log.Error("Python script stderr: ", stderr.String())
				return err
			}
			log.Info("updateWordClouds completed. Stdout: ", stdout.String())
		}

		return nil
	}))

	return job, err
}

func updateRedditPostsJob(h handlers.Handlers, scheduler gocron.Scheduler, atTimes gocron.AtTimes) (gocron.Job, error) {
	job, err := scheduler.NewJob(gocron.DailyJob(1, atTimes), gocron.NewTask(func() error {
		log.Info("Running updateRedditPostsJob")

		if err := h.UpdatePostsFromReddit(); err != nil {
			return err
		}

		log.Info("updateRedditPostsJob completed")
		return nil
	}))

	return job, err
}
