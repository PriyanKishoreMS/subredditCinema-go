package api

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/priyankishorems/bollytics-go/api/handlers"
	"golang.org/x/time/rate"
)

func IPRateLimit(h *handlers.Handlers) echo.MiddlewareFunc {

	type client struct {
		limiter  *rate.Limiter
		lastseen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// background routine to remove old entries from the map
	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastseen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			if h.Config.RateLimiter.Enabled {
				ip, _, err := net.SplitHostPort(c.Request().RemoteAddr)
				if err != nil {
					h.Utils.InternalServerError(c, err)
					return err
				}

				mu.Lock()

				_, found := clients[ip]
				if !found {
					clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(h.Config.RateLimiter.Rps), h.Config.RateLimiter.Burst)}
				}

				clients[ip].lastseen = time.Now()

				if !clients[ip].limiter.Allow() {
					mu.Unlock()
					h.Utils.RateLimitExceededResponse(c)
					return errors.New("rate limit exceeded")
				}

				mu.Unlock()
			}

			return next(c)
		}
	}
}
