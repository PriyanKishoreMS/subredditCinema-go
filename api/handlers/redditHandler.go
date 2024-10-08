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

func (h *Handlers) VerifySession(c echo.Context) error {
	reddit_id := c.Get("reddit_id").(string)
	return c.JSON(http.StatusOK, Cake{"message": "Session verified", "reddit_id": reddit_id})
}

func (h *Handlers) GetTrendingWordsHandler(sub string, interval string) ([]WordCount, error) {

	if slices.Index(subReddits, sub) == -1 {
		return nil, fmt.Errorf("invalid sub")
	}

	if slices.Index(intervals, interval) == -1 {
		return nil, fmt.Errorf("invalid interval")
	}
	var intervalInt int

	if interval == intervalWeek {
		intervalInt = 7
	} else {
		intervalInt = 30
	}

	allWords, err := h.Data.Posts.GetTrendingWords(sub, intervalInt)
	if err != nil {
		return nil, fmt.Errorf("error getting trending words %v", err)
	}

	trendingWords, err := h.getMostUsedWords(allWords, 100)
	if err != nil {
		return nil, fmt.Errorf("error getting most used words %v", err)
	}

	return trendingWords, nil
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

	frequency, err := h.Data.Posts.GetPostFrequency(sub)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("error getting post frequency %v", err)
	}

	frequencyMap, err := StructurePostFrequency(frequency)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("error structuring post frequency %v", err)
	}

	return c.JSON(http.StatusOK, Cake{fmt.Sprintf("%s_month_frequency", sub): frequencyMap})
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
		return c.JSON(http.StatusOK, Cake{"posts": []data.TopPosts{}})
	}

	return c.JSON(http.StatusOK, Cake{"posts": topPosts})
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
		return c.JSON(http.StatusOK, Cake{"message": "No users found"})
	}

	return c.JSON(http.StatusOK, Cake{"users": topUsers})
}

func (h *Handlers) UpdatePostsFromRedditHandler(c echo.Context) error {
	topPosts, err := GetDailyTopPosts(h)
	if err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("error getting top posts %v", err))
		return err
	}

	controversialPosts, err := GetDailyControversialPosts(h)
	if err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("error getting controversial posts %v", err))
		return err
	}

	allPosts := append(topPosts, controversialPosts...)

	if err = h.Data.Posts.InsertDailyPosts(allPosts); err != nil {
		h.Utils.InternalServerError(c, err)
		return err
	}

	return c.JSON(http.StatusOK, Cake{"message": "Posts updated successfully"})
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

	if err = h.Data.Posts.InsertDailyPosts(allPosts); err != nil {
		return err
	}

	fmt.Println("Posts updated successfully")
	return nil
}

func GetDailyTopPosts(h *Handlers) ([]data.Post, error) {
	var allPosts []data.Post

	dayPosts, err := getTopFromReddit(h.Reddit, 10, "day", allPosts)
	if err != nil {
		return nil, err
	}

	allPosts = append(allPosts, dayPosts...)

	weekPosts, err := getTopFromReddit(h.Reddit, 10, "week", allPosts)
	if err != nil {
		return nil, err
	}
	allPosts = append(allPosts, weekPosts...)

	monthPosts, err := getTopFromReddit(h.Reddit, 5, "month", allPosts)
	if err != nil {
		return nil, err
	}
	allPosts = append(allPosts, monthPosts...)

	return allPosts, nil
}

func GetDailyControversialPosts(h *Handlers) ([]data.Post, error) {
	var allPosts []data.Post

	dayPosts, err := getControversialFromReddit(h.Reddit, 10, "day", allPosts)
	if err != nil {
		return nil, err
	}
	allPosts = append(allPosts, dayPosts...)

	weekPosts, err := getControversialFromReddit(h.Reddit, 5, "week", allPosts)
	if err != nil {
		return nil, err
	}

	allPosts = append(allPosts, weekPosts...)

	monthPosts, err := getControversialFromReddit(h.Reddit, 5, "month", allPosts)
	if err != nil {
		return nil, err
	}
	allPosts = append(allPosts, monthPosts...)

	return allPosts, nil
}

func getTopFromReddit(Reddit *reddit.Client, limit int, interval string, allPosts []data.Post) ([]data.Post, error) {
	for _, sub := range subReddits {
		posts, _, err := Reddit.Subreddit.TopPosts(context.Background(), sub, &reddit.ListPostOptions{
			ListOptions: reddit.ListOptions{
				Limit: limit,
			},
			Time: interval,
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
				Category:             categoryTop,
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

func getControversialFromReddit(Reddit *reddit.Client, limit int, interval string, allPosts []data.Post) ([]data.Post, error) {
	for _, sub := range subReddits {
		posts, _, err := Reddit.Subreddit.ControversialPosts(context.Background(), sub, &reddit.ListPostOptions{
			ListOptions: reddit.ListOptions{
				Limit: limit,
			},
			Time: interval,
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
	}
	return allPosts, nil
}
