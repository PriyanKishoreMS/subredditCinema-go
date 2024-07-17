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
	User         string `json:"user"`
	TopPostCount int    `json:"post_count"`
}

func (p PostModel) GetTopUser(sub string, category string, interval int) ([]TopUsers, error) {
	ctx, cancel := Handlectx()
	defer cancel()
	var query string

	if category == "top" {
		query = `select author as user, count(*) as author_count from reddit_posts where subreddit=$1 and (category='top' or (category='controversial' and top_and_controversial=true)) and created_utc > now() - make_interval(days := $2) group by author order by author_count desc limit 5`
	} else if category == "controversial" {
		query = `select author as user, count(*) as author_count from reddit_posts where subreddit=$1 and (category='controversial' or (category='top' and top_and_controversial=true)) and created_utc > now() - make_interval(days := $2) group by author order by author_count desc limit 5`
	}

	rows, err := p.DB.Query(ctx, query, sub, interval)
	if err != nil {
		return nil, fmt.Errorf("error in getting top users; %v", err)
	}

	var topUsers []TopUsers
	for rows.Next() {
		var topUser TopUsers
		err = rows.Scan(&topUser.User, &topUser.TopPostCount)
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

	InsertPostsQuery := `INSERT INTO reddit_posts
	(id, name, created_utc, permalink, title, category, selftext, score, upvote_ratio, num_comments, subreddit, subreddit_id, subreddit_subscribers, author, author_fullname)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	ON CONFLICT(id) DO UPDATE
		SET top_and_controversial=TRUE
	WHERE reddit_posts.category <> EXCLUDED.category
	`

	for _, post := range dailyPosts {
		_, err := tx.Exec(ctx, InsertPostsQuery, post.ID, post.Name, post.CreatedUTC, post.Permalink, post.Title, post.Category, post.Selftext, post.Score, post.UpvoteRatio, post.NumComments, post.Subreddit, post.SubredditID, post.SubredditSubscribers, post.Author, post.AuthorFullname)
		if err != nil {
			return fmt.Errorf("error in inserting post: %v", err)
		}
	}

	DeleteOldPostsQuery := `DELETE FROM reddit_posts WHERE created_utc < NOW() - INTERVAL '30 days'`

	deleted, err := tx.Exec(ctx, DeleteOldPostsQuery)
	if err != nil {
		return fmt.Errorf("error in deleting old posts: %v", err)
	}

	fmt.Println("Deleted old posts: ", deleted.RowsAffected())

	return nil
}
