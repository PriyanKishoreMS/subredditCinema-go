package utils

import (
	"net"
	"net/http"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/oauth2"
)

var (
	DBName             string = os.Getenv("DB_DATABASE")
	DBUsername         string = os.Getenv("DB_USERNAME")
	DBPassword         string = os.Getenv("DB_PASSWORD")
	DBPort             string = os.Getenv("DB_PORT")
	DBHost             string = os.Getenv("DB_HOST")
	TMDBKey            string = os.Getenv("TMDB_API_KEY")
	RedditId           string = os.Getenv("REDDIT_API_ID")
	RedditSecret       string = os.Getenv("REDDIT_API_SECRET")
	RedditUsername     string = os.Getenv("REDDIT_USERNAME")
	RedditPassword     string = os.Getenv("REDDIT_PASSWORD")
	RedirectURL        string = "http://localhost:3000/callback"
	RedditUserAgent    string = os.Getenv("REDDIT_USER_AGENT")
	RedditIdWeb        string = os.Getenv("REDDIT_API_ID_WEB")
	RedditSecretWeb    string = os.Getenv("REDDIT_API_SECRET_WEB")
	RedditUserAgentWeb string = os.Getenv("REDDIT_USER_AGENT_WEB")
)

var (
	OauthConfig = &oauth2.Config{
		ClientID:     RedditIdWeb,
		ClientSecret: RedditSecretWeb,
		RedirectURL:  RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.reddit.com/api/v1/authorize",
			TokenURL: "https://www.reddit.com/api/v1/access_token",
		},
		Scopes: []string{"identity", "read"},
	}
)

var HttpClientConfig = &http.Client{
	Timeout: time.Second * 30,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

type Config struct {
	Port        int
	Env         string
	RateLimiter struct {
		Rps     int
		Burst   int
		Enabled bool
	}
}
