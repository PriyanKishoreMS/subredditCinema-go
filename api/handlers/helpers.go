package handlers

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/priyankishorems/bollytics-go/internal/data"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

func (h *Handlers) InsertFromJson(c echo.Context) error {
	files := []string{"sortedTopKollywood.json", "sortedControversialKollywood.json", "sortedTopMollywood.json", "sortedControversialMollywood.json", "sortedTopTollywood.json", "sortedControversialTollywood.json", "sortedTopBollywood.json", "sortedControversialBollywood.json"}

	for _, file := range files {
		err := h.Data.Posts.DumpJson(file)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Cake{"error": err.Error()})
		}
	}
	return c.JSON(http.StatusOK, Cake{"message": "Inserted from json completed broooo"})
}

func GetFromReddit(h *Handlers, c echo.Context) error {

	posts, _, err := h.Reddit.Subreddit.TopPosts(context.Background(), "MalayalamMovies", &reddit.ListPostOptions{
		ListOptions: reddit.ListOptions{
			Limit: 100,
			// After: "t3_1dqg62m",
		},
		Time: "month",
	})
	if err != nil {
		return err
	}
	return c.JSON(200, posts)
}

type PostFrequency struct {
	Hour  int
	Day   int
	Count int
}

func StructurePostFrequency(postFrequency []data.PostFrequency) (map[int][]int, error) {
	daysOfWeek := []int{0, 1, 2, 3, 4, 5, 6}
	postFrequencyMap := make(map[int][]int)

	for _, day := range daysOfWeek {
		postFrequencyMap[day] = make([]int, 24)
	}

	for _, pf := range postFrequency {
		day := pf.Day
		postFrequencyMap[day][pf.Hour] = pf.Count
	}

	return postFrequencyMap, nil
}
