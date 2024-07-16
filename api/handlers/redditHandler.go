package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

func (h *Handlers) RedditHomeHandler(c echo.Context) error {
	err := GetFromReddit(h, c)
	// err := DumpPosts(h, c)
	return err
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

func DumpPosts(h *Handlers, c echo.Context) error {
	err := h.Data.Posts.DumpJson("sortedControversialMollywood.json")
	if err != nil {
		fmt.Printf("error in dumping json; %v", err)
		return c.JSON(http.StatusInternalServerError, Cake{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, Cake{"message": "Dumped json"})
}
