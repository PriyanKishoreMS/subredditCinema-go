package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
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
func (h *Handlers) TimePerReq(c echo.Context) error {
	timeNow := time.Now()

	topPosts, err := GetDailyTopPosts(h)
	if err != nil {
		return err
	}

	timeAfer := time.Now()

	timeDiff := timeAfer.Sub(timeNow).Seconds()

	return c.JSON(http.StatusOK, Cake{"time": timeDiff, "posts": topPosts})
}

func (h *Handlers) ScaleData(c echo.Context) error {
	var allPosts []data.Post
	var after string

	for i := 0; i < 36; i++ {
		posts, resp, err := h.Reddit.Subreddit.ControversialPosts(context.Background(), "MalayalamMovies", &reddit.ListPostOptions{
			ListOptions: reddit.ListOptions{
				Limit: 100,
				After: after,
			},
			Time: "year",
		})
		if err != nil {
			h.Utils.InternalServerError(c, err)
			return fmt.Errorf("error getting top posts %v", err)
		}

		for _, post := range posts {
			allPosts = append(allPosts, data.Post{
				ID:                   post.ID,
				Name:                 post.FullID,
				CreatedUTC:           post.Created.Time,
				Permalink:            post.Permalink,
				Title:                post.Title,
				Category:             categoryControversial,
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
		log.Info(i, "th iteration with ", len(allPosts), " posts")
		after = resp.After
		log.Info("Inserting into db ", len(allPosts))

		err = h.Data.Posts.InsertDailyPosts(allPosts)
		if err != nil {
			h.Utils.InternalServerError(c, err)
			return fmt.Errorf("error inserting posts %v", err)
		}

		allPosts = nil
	}

	log.Info("Posts inserted successfully")

	return c.JSON(http.StatusOK, Cake{"posts": "inserted"})
}
