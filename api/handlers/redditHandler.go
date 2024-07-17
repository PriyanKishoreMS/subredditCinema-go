package handlers

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"
	"github.com/priyankishorems/bollytics-go/internal/data"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

var subReddits []string = []string{
	"kollywood", "MalayalamMovies", "tollywood", "bollywood",
}

const (
	intervalWeek          = "week"
	intervalMonth         = "month"
	categoryTop           = "top"
	categoryControversial = "controversial"
)

func (h *Handlers) GetTopUsers(c echo.Context) error {

	sub, err := h.Utils.ReadStringParam(c, "sub")
	if err != nil {
		h.Utils.BadRequest(c, err)
		return fmt.Errorf("invalid sub %v", err)
	}

	if slices.Index(subReddits, sub) == -1 {
		h.Utils.BadRequest(c, fmt.Errorf("invalid sub"))
		return fmt.Errorf("invalid sub")
	}

	interval, err := h.Utils.ReadStringParam(c, "interval")
	if err != nil {
		h.Utils.BadRequest(c, err)
		return fmt.Errorf("invalid interval %v", err)
	}

	if interval != intervalWeek && interval != intervalMonth {
		h.Utils.BadRequest(c, fmt.Errorf("invalid interval"))
		return fmt.Errorf("invalid interval")
	}

	category, err := h.Utils.ReadStringParam(c, "category")
	if err != nil {
		h.Utils.BadRequest(c, err)
		return fmt.Errorf("invalid category %v", err)

	}

	if category != categoryTop && category != categoryControversial {
		h.Utils.BadRequest(c, fmt.Errorf("invalid category"))
		return fmt.Errorf("invalid category")
	}

	var intervalInt int

	if interval == intervalWeek {
		intervalInt = 7
	} else {
		intervalInt = 30
	}

	topUsers, err := h.Data.Posts.GetTopUser(sub, category, intervalInt)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("error getting top users %v", err)
	}

	return c.JSON(http.StatusOK, Cake{fmt.Sprintf("%s_%s_%s_users", sub, category, interval): topUsers})
}

func (h *Handlers) UpdatePostsFromReddit() error {
	topPosts, err := GetDailyTopPosts(h)
	if err != nil {
		return err
	}

	controversialPosts, err := GetDailyControversialPosts(h)
	if err != nil {
		return err
	}

	allPosts := append(topPosts, controversialPosts...)

	err = h.Data.Posts.InsertDailyPosts(allPosts)
	if err != nil {
		return err
	}

	fmt.Println("Posts updated successfully")
	return nil
}

func GetDailyTopPosts(h *Handlers) ([]data.Post, error) {
	var allPosts []data.Post

	for _, sub := range subReddits {
		posts, _, err := h.Reddit.Subreddit.TopPosts(context.Background(), sub, &reddit.ListPostOptions{
			ListOptions: reddit.ListOptions{
				Limit: 10,
			},
			Time: "day",
		})
		if err != nil {
			return nil, err
		}

		for _, post := range posts {
			allPosts = append(allPosts, data.Post{
				ID:                   post.ID,
				Name:                 post.FullID,
				CreatedUTC:           post.Created.Time,
				Permalink:            post.Permalink,
				Title:                post.Title,
				Category:             "top",
				Selftext:             post.Body,
				Score:                post.Score,
				UpvoteRatio:          float64(post.UpvoteRatio),
				NumComments:          post.NumberOfComments,
				Subreddit:            post.SubredditName,
				SubredditID:          post.SubredditID,
				SubredditSubscribers: post.SubredditSubscribers,
				Author:               post.Author,
				AuthorFullname:       post.AuthorID,
			})

		}
	}

	return allPosts, nil
}

func GetDailyControversialPosts(h *Handlers) ([]data.Post, error) {
	var allPosts []data.Post

	for _, sub := range subReddits {
		posts, _, err := h.Reddit.Subreddit.ControversialPosts(context.Background(), sub, &reddit.ListPostOptions{
			ListOptions: reddit.ListOptions{
				Limit: 10,
			},
			Time: "day",
		})
		if err != nil {
			return nil, err
		}

		for _, post := range posts {
			allPosts = append(allPosts, data.Post{
				ID:                   post.ID,
				Name:                 post.FullID,
				CreatedUTC:           post.Created.Time,
				Permalink:            post.Permalink,
				Title:                post.Title,
				Category:             "controversial",
				Selftext:             post.Body,
				Score:                post.Score,
				UpvoteRatio:          float64(post.UpvoteRatio),
				NumComments:          post.NumberOfComments,
				Subreddit:            post.SubredditName,
				SubredditID:          post.SubredditID,
				SubredditSubscribers: post.SubredditSubscribers,
				Author:               post.Author,
				AuthorFullname:       post.AuthorID,
			})

		}
	}

	return allPosts, nil
}
