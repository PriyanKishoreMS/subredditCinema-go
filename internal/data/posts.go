package data

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	pgx "github.com/jackc/pgx/v5/pgxpool"
)

type PostModel struct {
	DB *pgx.Pool
}

type Post struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	CreatedUTC           time.Time `json:"created_utc"`
	Permalink            string    `json:"permalink"`
	Title                string    `json:"title"`
	Category             string    `json:"category"`
	Selftext             string    `json:"selftext"`
	Score                int       `json:"score"`
	UpvoteRatio          float64   `json:"upvote_ratio"`
	NumComments          int       `json:"num_comments"`
	Subreddit            string    `json:"subreddit"`
	SubredditID          string    `json:"subreddit_id"`
	SubredditSubscribers int       `json:"subreddit_subscribers"`
	Author               string    `json:"author"`
	AuthorFullname       string    `json:"author_fullname"`
}

type PostsWrapper struct {
	Posts []Post `json:"posts"`
}
type TopUsers struct {
	User      string `json:"user"`
	PostCount int    `json:"post_count"`
}

type TopPosts struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Body          string  `json:"body"`
	Author        string  `json:"author"`
	URL           string  `json:"url"`
	Upvotes       int     `json:"upvotes"`
	UpvoteRatio   float64 `json:"upvote_ratio"`
	Subreddit     string  `json:"subreddit"`
	NumComments   int     `json:"num_comments"`
	Category      string  `json:"category"`
	CategoryScore float64 `json:"category_score"`
}

type PostFrequency struct {
	Hour  int
	Day   int
	Count int
}

func (p PostModel) GetTrendingWords(sub string, interval int) ([]string, error) {
	ctx, cancel := Handlectx()
	defer cancel()

	query := GetAllTextsOfInterval

	rows, err := p.DB.Query(ctx, query, sub, interval)
	if err != nil {
		return nil, fmt.Errorf("error in getting trending words; %v", err)
	}
	defer rows.Close()

	var words []string
	for rows.Next() {
		var word string
		err = rows.Scan(&word)
		if err != nil {
			return nil, fmt.Errorf("error in scanning trending words; %v", err)
		}
		words = append(words, word)
	}

	return words, nil
}

func (p PostModel) GetPostFrequency(sub string, interval int) ([]PostFrequency, error) {
	ctx, cancel := Handlectx()
	defer cancel()

	query := FrequencyOfPostsQuery

	rows, err := p.DB.Query(ctx, query, sub, interval)
	if err != nil {
		return nil, fmt.Errorf("error in getting post frequency by day of week; %v", err)
	}
	defer rows.Close()

	var postFrequency []PostFrequency
	for rows.Next() {
		var frequency PostFrequency
		err = rows.Scan(&frequency.Hour, &frequency.Day, &frequency.Count)
		if err != nil {
			return nil, fmt.Errorf("error in scanning post frequency; %v", err)
		}
		postFrequency = append(postFrequency, frequency)
	}
	return postFrequency, nil
}

func (p PostModel) GetTopPosts(sub string, category string, interval int) ([]TopPosts, error) {
	ctx, cancel := Handlectx()
	defer cancel()

	var query string

	switch category {
	case "top":
		query = TopPostsQuery
	case "controversial":
		query = ControversialPostsQuery
	case "top_and_controversial":
		query = TopAndControversialPostsQuery
	case "hated":
		query = MostHatedPostsQuery
	default:
		return nil, fmt.Errorf("invalid category: %s", category)
	}

	rows, err := p.DB.Query(ctx, query, sub, interval)
	if err != nil {
		return nil, fmt.Errorf("error in getting top posts; %v", err)
	}
	defer rows.Close()

	var topPosts []TopPosts
	for rows.Next() {
		var topPost TopPosts
		err = rows.Scan(&topPost.ID, &topPost.Title, &topPost.Body, &topPost.Author, &topPost.URL, &topPost.Upvotes, &topPost.UpvoteRatio, &topPost.Subreddit, &topPost.NumComments, &topPost.Category, &topPost.CategoryScore)
		if err != nil {
			return nil, fmt.Errorf("error in scanning top posts; %v", err)
		}
		topPosts = append(topPosts, topPost)
	}

	return topPosts, nil
}

func (p PostModel) GetTopUser(sub string, category string, interval int) ([]TopUsers, error) {
	ctx, cancel := Handlectx()
	defer cancel()
	var query string

	switch category {
	case "top":
		query = TopUsersQuery
	case "controversial":
		query = ControversialUsersQuery
	default:
		return nil, fmt.Errorf("invalid category: %s", category)
	}

	rows, err := p.DB.Query(ctx, query, sub, interval)
	if err != nil {
		return nil, fmt.Errorf("error in getting top users; %v", err)
	}
	defer rows.Close()

	var topUsers []TopUsers
	for rows.Next() {
		var topUser TopUsers
		err = rows.Scan(&topUser.User, &topUser.PostCount)
		if err != nil {
			return nil, fmt.Errorf("error in scanning top users; %v", err)
		}
		topUsers = append(topUsers, topUser)
	}

	return topUsers, nil
}

func (p PostModel) DumpJson(filename string) error {
	fullpath := filepath.Join("dump/", filename)
	jsonFile, err := os.Open(fullpath)
	if err != nil {
		return fmt.Errorf("error in opening json file; %v", err)
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return fmt.Errorf("error in reading json file; %v", err)
	}

	var postsWrapper PostsWrapper
	err = json.Unmarshal(byteValue, &postsWrapper)
	if err != nil {
		return fmt.Errorf("error in unmarshalling json; %v", err)
	}

	err = p.InsertDailyPosts(postsWrapper.Posts)
	if err != nil {
		return fmt.Errorf("error in inserting daily posts; %v", err)
	}

	return nil
}

func (p PostModel) InsertDailyPosts(dailyPosts []Post) error {
	ctx, cancel := Handlectx()
	defer cancel()

	tx, err := p.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error in starting transaction; %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(ctx)
			err = fmt.Errorf("transaction panicked: %v", r)
		} else if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	query := InsertPostsQuery

	for _, post := range dailyPosts {
		_, err := tx.Exec(ctx, query, post.ID, post.Name, post.CreatedUTC, post.Permalink, post.Title, post.Category, post.Selftext, post.Score, post.UpvoteRatio, post.NumComments, post.Subreddit, post.SubredditID, post.SubredditSubscribers, post.Author, post.AuthorFullname)
		if err != nil {
			return fmt.Errorf("error in inserting post: %v", err)
		}
	}

	query = DeleteOldPostsQuery

	deleted, err := tx.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("error in deleting old posts: %v", err)
	}

	fmt.Println("Deleted old posts: ", deleted.RowsAffected())

	return nil
}
