package handlers

import (
	"encoding/json"
	"fmt"
	"io"
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

var excludedWords []string = []string{"movie", "movies", "watch", "film", "time", "films", "like", "watching", "good", "seen", "watched", "best", "better", "love", "loved", "https", "http", "webp", "png", "scene", "scenes", "song", "songs", "post", "posts", "guy", "guys", "people", "tamil", "telugu", "hindi", "malayalam", "kollywood", "bollywood", "mollywood", "tollywood", "music", "story", "actor", "actors", "youtube", "cinema", "release", "youtu", "instagram", "kinda", "share", "character", "characters", "video", "screen", "content", "version", "industry", "reddit"}

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

type UserType struct {
	RedditID string `json:"reddit_id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
}

func (h *Handlers) GetAuthUserDataFromReddit(c echo.Context, token *oauth2.Token, userAgent string) (*UserType, error) {

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

	var userData UserType = UserType{
		RedditID: user["id"].(string),
		Name:     user["name"].(string),
		Avatar:   user["snoovatar_img"].(string),
	}

	return &userData, nil
}

func (h *Handlers) GetFromReddit(c echo.Context) error {

	posts, _, err := h.Reddit.Subreddit.SearchPosts(c.Request().Context(), "Average Tamil Cinema", "kollywood", &reddit.ListPostSearchOptions{
		ListPostOptions: reddit.ListPostOptions{
			Time: "year",
		},
		Sort: "top",
	})
	if err != nil {
		log.Error("Error getting posts from reddit", err)
		return c.JSON(http.StatusInternalServerError, Cake{"error": err.Error()})
	}

	index := 0

	input := data.Post{
		ID:                   posts[index].ID,
		Name:                 posts[index].FullID,
		CreatedUTC:           posts[index].Created.Time,
		Permalink:            posts[index].Permalink,
		Title:                posts[index].Title,
		Category:             "top",
		Selftext:             posts[index].Body,
		Score:                posts[index].Score,
		UpvoteRatio:          float64(posts[index].UpvoteRatio),
		NumComments:          posts[index].NumberOfComments,
		Subreddit:            posts[index].SubredditName,
		SubredditID:          posts[index].SubredditID,
		SubredditSubscribers: posts[index].SubredditSubscribers,
		Author:               posts[index].Author,
		AuthorFullname:       posts[index].AuthorID,
	}

	err = h.Data.Posts.InsertOnePost(input)
	if err != nil {
		log.Error("Error inserting post into db", err)
		return c.JSON(http.StatusInternalServerError, Cake{"error": err.Error()})
	}

	return c.JSON(200, posts[0])
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

func (h *Handlers) GetTrendingWordsHandlerWeb(c echo.Context) error {
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

	trendingWords, err := h.getMostUsedWords(allWords, 100)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("error getting most used words %v", err)
	}

	return c.JSON(http.StatusOK, Cake{fmt.Sprintf("%s_%s_trending_words", sub, interval): trendingWords})
}

// reddit api doesn't provide snoovatar data. This url of reddit.com/user/{username}/about.json provides snoovatar data
// but it is inconsistent, doesn't provide proper data for most users, also the image is not png, so not using it.
func (h *Handlers) GetRedditUsersSnoovatar(c echo.Context, topUsers []data.TopUsers) error {
	for i, _ := range topUsers {
		user := &topUsers[i]

		httpClient := http.Client{}
		url := fmt.Sprintf("https://www.reddit.com/user/%s/about.json", user.User)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}

		if err != nil {
			h.Utils.InternalServerError(c, err)
			return err
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var userData struct {
			Data struct {
				SnoovatarImg string `json:"snoovatar_img"`
			}
		}

		err = json.Unmarshal(body, &userData)
		if err != nil {
			return err
		}

		user.Avatar = userData.Data.SnoovatarImg
	}
	return nil
}
