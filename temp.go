package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

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

func main() {
	jsonFile, err := os.Open("topMollywood.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	var postsWrapper PostsWrapper
	err = json.Unmarshal(byteValue, &postsWrapper)
	if err != nil {
		fmt.Println(err)
		return
	}

	AddCategory(&postsWrapper, "top")
	SortPosts(postsWrapper)

	count := 1
	date := postsWrapper.Posts[0].CreatedUTC
	for _, post := range postsWrapper.Posts {
		createdDate := post.CreatedUTC.Day()
		dateDate := date.Day()
		if createdDate != dateDate {
			fmt.Printf("%s. %d\n", post.CreatedUTC, count)
			date = post.CreatedUTC
			count = 1
		}
		count++
	}
}

func AddCategory(postsWrapper *PostsWrapper, category string) {
	for i := range postsWrapper.Posts {
		postsWrapper.Posts[i].Category = category
	}
}

func SortPosts(postsWrapper PostsWrapper) {
	sort.Slice(postsWrapper.Posts, func(i, j int) bool {
		dateI := postsWrapper.Posts[i].CreatedUTC.Truncate(24 * time.Hour)
		dateJ := postsWrapper.Posts[j].CreatedUTC.Truncate(24 * time.Hour)

		if dateI.Equal(dateJ) {
			return postsWrapper.Posts[i].Score > postsWrapper.Posts[j].Score
		}
		return dateI.After(dateJ)
	})

	sortedFile, err := os.Create("sortedTopMollywood.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sortedFile.Close()

	sortedBytes, err := json.MarshalIndent(postsWrapper, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = sortedFile.Write(sortedBytes)
	if err != nil {
		fmt.Println(err)
		return
	}
}
