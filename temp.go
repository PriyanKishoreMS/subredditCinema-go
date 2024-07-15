package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

// Post represents the structure of a single post
type Post struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	CreatedUTC           time.Time `json:"created_utc"`
	Permalink            string    `json:"permalink"`
	Title                string    `json:"title"`
	Score                int       `json:"score"`
	UpvoteRatio          float64   `json:"upvote_ratio"`
	NumComments          int       `json:"num_comments"`
	Subreddit            string    `json:"subreddit"`
	SubredditID          string    `json:"subreddit_id"`
	SubredditSubscribers int       `json:"subreddit_subscribers"`
	Author               string    `json:"author"`
	Over18               bool      `json:"over_18"`
}

// Wrapper struct to match the JSON structure
type PostsWrapper struct {
	Posts []Post `json:"posts"`
}

func main() {
	jsonFile, err := os.Open("sortedTopKollywood.json")
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

	// SortPosts(postsWrapper)

	count := 1
	date := postsWrapper.Posts[0].CreatedUTC
	for _, post := range postsWrapper.Posts {
		createdDate := post.CreatedUTC.Day()
		dateDate := date.Day()
		if createdDate != dateDate {
			date = post.CreatedUTC
			fmt.Printf("%s. %d\n", post.CreatedUTC, count)
			count = 1
		}
		count++
	}
}

func SortPosts(postsWrapper PostsWrapper) {
	sort.Slice(postsWrapper.Posts, func(i, j int) bool {
		return postsWrapper.Posts[j].CreatedUTC.Before(postsWrapper.Posts[i].CreatedUTC)
	})

	sortedFile, err := os.Create("sortedTopKollywood.json")
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
