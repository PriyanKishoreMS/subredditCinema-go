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
	intervalWeek                = "week"
	intervalMonth               = "month"
	interval6Months             = "6months"
	intervalYear                = "year"
	categoryTop                 = "top"
	categoryControversial       = "controversial"
	categoryTopAndControversial = "top_and_controversial"
	categoryHated               = "hated"
)

var intervals = []string{intervalWeek, intervalMonth, interval6Months, intervalYear}

func (h *Handlers) GetTrendingWordsHandler(c echo.Context) error {
	sub, err := h.Utils.ReadStringParam(c, "sub")
	if err != nil {
		h.Utils.BadRequest(c, err)
		return fmt.Errorf("invalid sub %v", err)
	}

	if slices.Index(subReddits, sub) == -1 {
		h.Utils.BadRequest(c, fmt.Errorf("invalid sub"))
		return fmt.Errorf("invalid sub")
	}

	interval := h.Utils.ReadStringQuery(c.QueryParams(), "interval", intervalMonth)

	if slices.Index(intervals, interval) == -1 {
		fmt.Println(interval, "interval")
		h.Utils.BadRequest(c, fmt.Errorf("invalid interval"))
		return fmt.Errorf("invalid interval")
	}
	var intervalInt int

	if interval == intervalWeek {
		intervalInt = 7
	} else {
		intervalInt = 30
	}

	allWords, err := h.Data.Posts.GetTrendingWords(sub, intervalInt)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("error getting trending words %v", err)
	}

	trendingWords, err := getMostUsedWords(allWords, 100)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("error getting most used words %v", err)
	}

	return c.JSON(http.StatusOK, Cake{fmt.Sprintf("%s_%s_trending_words", sub, interval): trendingWords})
}

func (h *Handlers) GetPostFrequencyHandler(c echo.Context) error {
	sub, err := h.Utils.ReadStringParam(c, "sub")
	if err != nil {
		h.Utils.BadRequest(c, err)
		return fmt.Errorf("invalid sub %v", err)
	}

	if slices.Index(subReddits, sub) == -1 {
		h.Utils.BadRequest(c, fmt.Errorf("invalid sub"))
		return fmt.Errorf("invalid sub")
	}

	interval := h.Utils.ReadStringQuery(c.QueryParams(), "interval", intervalMonth)

	if slices.Index(intervals, interval) == -1 {
		fmt.Println(interval, "interval")
		h.Utils.BadRequest(c, fmt.Errorf("invalid interval"))
		return fmt.Errorf("invalid interval")
	}
	var intervalInt int

	if interval == intervalWeek {
		intervalInt = 7
	} else if interval == intervalMonth {
		intervalInt = 30
	} else if interval == interval6Months {
		intervalInt = 180
	} else {
		intervalInt = 365
	}

	frequency, err := h.Data.Posts.GetPostFrequency(sub, intervalInt)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("error getting post frequency %v", err)
	}

	frequencyMap, err := StructurePostFrequency(frequency)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("error structuring post frequency %v", err)
	}

	return c.JSON(http.StatusOK, Cake{fmt.Sprintf("%s_%s_frequency", sub, interval): frequencyMap})
}

func (h *Handlers) GetTopPostsHandler(c echo.Context) error {

	sub, err := h.Utils.ReadStringParam(c, "sub")
	if err != nil {
		h.Utils.BadRequest(c, err)
		return fmt.Errorf("invalid sub %v", err)
	}

	if slices.Index(subReddits, sub) == -1 {
		h.Utils.BadRequest(c, fmt.Errorf("invalid sub"))
		return fmt.Errorf("invalid sub")
	}

	interval := h.Utils.ReadStringQuery(c.QueryParams(), "interval", intervalMonth)

	category, err := h.Utils.ReadStringParam(c, "category")
	if err != nil {
		h.Utils.BadRequest(c, err)
		return fmt.Errorf("invalid category %v", err)

	}

	if category != categoryTop && category != categoryControversial && category != categoryTopAndControversial && category != categoryHated {
		h.Utils.BadRequest(c, fmt.Errorf("invalid category"))
		return fmt.Errorf("invalid category")
	}

	if slices.Index(intervals, interval) == -1 {
		fmt.Println(interval, "interval")
		h.Utils.BadRequest(c, fmt.Errorf("invalid interval"))
		return fmt.Errorf("invalid interval")
	}

	var intervalInt int

	if interval == intervalWeek {
		intervalInt = 7
	} else if interval == intervalMonth {
		intervalInt = 30
	} else if interval == interval6Months {
		intervalInt = 180
	} else {
		intervalInt = 365
	}

	topPosts, err := h.Data.Posts.GetTopPosts(sub, category, intervalInt)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("error getting top users %v", err)
	}

	responseLength := len(topPosts)
	if responseLength < 1 {
		return c.JSON(http.StatusNotFound, Cake{"message": "No posts found"})
	}

	return c.JSON(http.StatusOK, Cake{fmt.Sprintf("%s_%s_%s_posts", sub, category, interval): topPosts})
}

func (h *Handlers) GetTopUsersHandler(c echo.Context) error {

	sub, err := h.Utils.ReadStringParam(c, "sub")
	if err != nil {
		h.Utils.BadRequest(c, err)
		return fmt.Errorf("invalid sub %v", err)
	}

	if slices.Index(subReddits, sub) == -1 {
		h.Utils.BadRequest(c, fmt.Errorf("invalid sub"))
		return fmt.Errorf("invalid sub")
	}

	interval := h.Utils.ReadStringQuery(c.QueryParams(), "interval", intervalMonth)

	category, err := h.Utils.ReadStringParam(c, "category")
	if err != nil {
		h.Utils.BadRequest(c, err)
		return fmt.Errorf("invalid category %v", err)

	}

	if category != categoryTop && category != categoryControversial {
		h.Utils.BadRequest(c, fmt.Errorf("invalid category"))
		return fmt.Errorf("invalid category")
	}

	if slices.Index(intervals, interval) == -1 {
		fmt.Println(interval, "interval")
		h.Utils.BadRequest(c, fmt.Errorf("invalid interval"))
		return fmt.Errorf("invalid interval")
	}

	var intervalInt int

	if interval == intervalWeek {
		intervalInt = 7
	} else if interval == intervalMonth {
		intervalInt = 30
	} else if interval == interval6Months {
		intervalInt = 180
	} else {
		intervalInt = 365
	}

	topUsers, err := h.Data.Posts.GetTopUser(sub, category, intervalInt)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("error getting top users %v", err)
	}

	responseLength := len(topUsers)
	if responseLength < 1 {
		return c.JSON(http.StatusNotFound, Cake{"message": "No posts found"})
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
