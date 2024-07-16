package handlers

import (
	"context"
	"fmt"

	"github.com/priyankishorems/bollytics-go/internal/data"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

var subReddits []string = []string{
	"kollywood", "MalayalamMovies", "tollywood", "bollywood",
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
