package api

import (
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/priyankishorems/bollytics-go/api/handlers"
	"golang.org/x/time/rate"
)

func AuthenticateUserSession(h *handlers.Handlers) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			reddit_id := h.SessionManager.GetString(ctx, "reddit_id")
			if reddit_id == "" {
				h.Utils.UserUnAuthorizedResponse(c)
				return errors.New("no session")
			}

			exists, err := h.Data.Users.CheckUserExists(reddit_id)
			if err != nil {
				h.Utils.InternalServerError(c, err)
				return err
			}

			if !exists {
				h.Utils.UserUnAuthorizedResponse(c)
				return errors.New("user not found")
			}

			c.Set("reddit_id", reddit_id)
			return next(c)
		}
	}
}

func ManageSession(h *handlers.Handlers) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			var token string
			cookie, err := c.Cookie(h.SessionManager.Cookie.Name)
			if err == nil {
				token = cookie.Value
			}

			ctx, err = h.SessionManager.Load(ctx, token)
			if err != nil {
				return err
			}

			c.SetRequest(c.Request().WithContext(ctx))

			c.Response().Before(func() {
				if h.SessionManager.Status(ctx) != scs.Unmodified {
					responseCookie := &http.Cookie{
						Name:     h.SessionManager.Cookie.Name,
						Path:     h.SessionManager.Cookie.Path,
						Domain:   h.SessionManager.Cookie.Domain,
						Secure:   h.SessionManager.Cookie.Secure,
						HttpOnly: h.SessionManager.Cookie.HttpOnly,
						SameSite: h.SessionManager.Cookie.SameSite,
					}

					switch h.SessionManager.Status(ctx) {
					case scs.Modified:
						token, _, err := h.SessionManager.Commit(ctx)
						if err != nil {
							log.Error("Failed to commit session: ", err)
						}

						responseCookie.Value = token

					case scs.Destroyed:
						responseCookie.Expires = time.Unix(1, 0)
						responseCookie.MaxAge = -1
					}

					c.SetCookie(responseCookie)
					h.Utils.AddHeaderIfMissing(c.Response(), "Cache-Control", `no-cache="Set-Cookie"`)
					h.Utils.AddHeaderIfMissing(c.Response(), "Vary", "Cookie")
				}
			})

			return next(c)
		}
	}
}

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
