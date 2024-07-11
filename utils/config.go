package utils

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	DBName     string = os.Getenv("DB_DATABASE")
	DBUsername string = os.Getenv("DB_USERNAME")
	DBPassword string = os.Getenv("DB_PASSWORD")
	DBPort     string = os.Getenv("DB_PORT")
	DBHost     string = os.Getenv("DB_HOST")
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
