package handlers

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

func (h *Handlers) RedditHomeHandler(c echo.Context) error {
	// harvest, err := h.RedditBot.Listing("/r/kollywood", "")
	// if err != nil {
	// 	return err
	// }

	// posts := []map[string]string{}

	// for _, post := range harvest.Posts[:5] {
	// 	posts = append(posts, map[string]string{
	// 		"author": post.Author,
	// 		"title":  post.Title,
	// 	})
	// }

	// return c.JSON(200, posts)

	posts, _, err := h.Reddit.Subreddit.TopPosts(context.Background(), "kollywood", &reddit.ListPostOptions{
		ListOptions: reddit.ListOptions{
			Limit: 100,
			After: "t3_1e28kya",
		},
		Time: "month",
	})
	if err != nil {
		return err
	}
	return c.JSON(200, posts)
}
