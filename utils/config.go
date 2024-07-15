package utils

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	DBName         string = os.Getenv("DB_DATABASE")
	DBUsername     string = os.Getenv("DB_USERNAME")
	DBPassword     string = os.Getenv("DB_PASSWORD")
	DBPort         string = os.Getenv("DB_PORT")
	DBHost         string = os.Getenv("DB_HOST")
	TMDBKey        string = os.Getenv("TMDB_API_KEY")
	RedditId       string = os.Getenv("REDDIT_API_ID")
	RedditSecret   string = os.Getenv("REDDIT_API_SECRET")
	RedditUsername string = os.Getenv("REDDIT_USERNAME")
	RedditPassword string = os.Getenv("REDDIT_PASSWORD")
)

type Config struct {
	Port        int
	Env         string
	RateLimiter struct {
		Rps     int
		Burst   int
		Enabled bool
	}
}
