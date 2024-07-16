package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
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

func DumpPosts(h *Handlers, c echo.Context, filename string) error {
	err := h.Data.Posts.DumpJson(filename)
	if err != nil {
		fmt.Printf("error in dumping json; %v", err)
		return c.JSON(http.StatusInternalServerError, Cake{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, Cake{"message": "Dumped json"})
}
