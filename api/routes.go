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

	// limiterStore := middleware.NewRateLimiterMemoryStore(20)
	// e.Use(middleware.RateLimiter(limiterStore))

	e.HideBanner = true
	e.GET("/", h.HomeFunc)
	e.GET("/actors/:name", h.SearchActorsHandler)
	e.GET("/movies/:name", h.SearchMoviesHandler)
	return e
}
