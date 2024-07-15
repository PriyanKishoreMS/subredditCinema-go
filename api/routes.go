package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/priyankishorems/bollytics-go/api/handlers"
)

func SetupRoutes(h *handlers.Handlers) *echo.Echo {
	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(IPRateLimit(h))
	e.Use(middleware.RemoveTrailingSlash())

	// limiterStore := middleware.NewRateLimiterMemoryStore(20)
	// e.Use(middleware.RateLimiter(limiterStore))

	e.HideBanner = true
	e.GET("/", h.HomeFunc)

	api := e.Group("/api")
	{
		api.GET("/actors/:name", h.SearchActorsHandler)
		api.GET("/movies/:name", h.SearchMoviesHandler)

		reddit := api.Group("/reddit")
		{
			reddit.GET("", h.RedditHomeHandler)
		}
	}

	return e
}
