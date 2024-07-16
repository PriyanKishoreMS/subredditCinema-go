package data

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	pgx "github.com/jackc/pgx/v5"
)

type PostModel struct {
	DB *pgx.Conn
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

func (t PostModel) CreateTask() error {
	return nil
}

func (t PostModel) DumpJson(filename string) error {
	jsonFile, err := os.Open(filename)
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

	ctx, cancel := Handlectx()
	defer cancel()

	tx, err := t.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error in starting transaction; %v", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	query := `INSERT INTO reddit_posts 
	(id, name, created_utc, permalink, title, category, selftext, score, upvote_ratio, num_comments, subreddit, subreddit_id, subreddit_subscribers, author, author_fullname) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	ON CONFLICT(id) DO UPDATE 
		SET top_and_controversial=TRUE	
	WHERE reddit_posts.category <> EXCLUDED.category
	`

	for _, post := range postsWrapper.Posts {
		_, err := tx.Exec(ctx, query, post.ID, post.Name, post.CreatedUTC, post.Permalink, post.Title, post.Category, post.Selftext, post.Score, post.UpvoteRatio, post.NumComments, post.Subreddit, post.SubredditID, post.SubredditSubscribers, post.Author, post.AuthorFullname)
		if err != nil {
			return fmt.Errorf("error in inserting post: %v", err)
		}
	}

	return nil
}
