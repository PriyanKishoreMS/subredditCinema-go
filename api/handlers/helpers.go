package handlers

import (
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/priyankishorems/bollytics-go/internal/data"
	"github.com/priyankishorems/bollytics-go/utils"
	"github.com/vartanbeno/go-reddit/v2/reddit"
	"golang.org/x/oauth2"
)

var excludedWords []string = []string{"movie", "movies", "watch", "film", "time", "films", "like", "watching", "good", "seen", "watched", "best", "better", "love", "loved", "https", "http", "webp", "png", "scene", "scenes", "song", "songs", "post", "posts", "guy", "guys", "people", "tamil", "telugu", "hindi", "malayalam", "kollywood", "bollywood", "mollywood", "tollywood", "music", "story", "actor", "actors"}

type WordCount struct {
	Word  string
	Count int
}

func (h *Handlers) getMostUsedWords(texts []string, limit int) ([]WordCount, error) {
	wordCounts := make(map[string]int)

	for _, text := range texts {
		cleanText := h.Stopword.ClearStringByLang(strings.ToLower(text), "en")

		words := strings.Fields(cleanText)
		for _, word := range words {
			if len(word) > 4 {
				if slices.Index(excludedWords, word) == -1 {
					wordCounts[word]++
				}
			}
		}
	}

	var wordCountSlice []WordCount
	for word, count := range wordCounts {
		wordCountSlice = append(wordCountSlice, WordCount{word, count})
	}

	sort.Slice(wordCountSlice, func(i, j int) bool {
		return wordCountSlice[i].Count > wordCountSlice[j].Count
	})

	if len(wordCountSlice) > limit {
		return wordCountSlice[:limit], nil
	}
	return wordCountSlice, nil
}

func (h *Handlers) GetAuthUserDataFromReddit(c echo.Context, token *oauth2.Token, userAgent string) (Cake, error) {

	httpClient := utils.OauthConfig.Client(c.Request().Context(), token)
	url := "https://oauth.reddit.com/api/v1/me"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	user, err := h.Utils.MakeCustomRequest(httpClient, req)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (h *Handlers) GetFromReddit(c echo.Context) error {

	posts, _, err := h.Reddit.Subreddit.SearchPosts(c.Request().Context(), "Bujji in Chennai", "kollywood", &reddit.ListPostSearchOptions{
		ListPostOptions: reddit.ListPostOptions{
			Time: "year",
		},
		Sort: "top",
	})
	if err != nil {
		log.Error("Error getting posts from reddit", err)
		return c.JSON(http.StatusInternalServerError, Cake{"error": err.Error()})
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
		posts, resp, err := h.Reddit.Subreddit.ControversialPosts(c.Request().Context(), "MalayalamMovies", &reddit.ListPostOptions{
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

// br0000 its moderator only
func (h *Handlers) GetTrafficHandler(c echo.Context) error {

	sub, err := h.Utils.ReadStringParam(c, "sub")
	if err != nil {
		h.Utils.BadRequest(c, err)
		return fmt.Errorf("invalid sub %v", err)
	}

	if slices.Index(subReddits, sub) == -1 {
		h.Utils.BadRequest(c, fmt.Errorf("invalid sub"))
		return fmt.Errorf("invalid sub")
	}

	day, hour, month, _, err := h.Reddit.Subreddit.Traffic(c.Request().Context(), sub)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("error getting traffic %v", err)
	}

	var traffic = Cake{
		"hour_traffic":  hour,
		"day_traffic":   day,
		"month_traffic": month,
	}

	return c.JSON(http.StatusOK, Cake{"traffic": traffic})
}
