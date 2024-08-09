package jobs

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/go-co-op/gocron/v2"
	"github.com/labstack/gommon/log"
	"github.com/priyankishorems/bollytics-go/api/handlers"
)

func UpdateWordClouds(h handlers.Handlers, scheduler gocron.Scheduler, atTimes gocron.AtTimes) (gocron.Job, error) {
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

func UpdateRedditPostsJob(h handlers.Handlers, scheduler gocron.Scheduler, atTimes gocron.AtTimes) (gocron.Job, error) {
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
