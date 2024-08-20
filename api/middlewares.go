package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pascaldekloe/jwt"
	"github.com/priyankishorems/bollytics-go/api/handlers"
	"github.com/priyankishorems/bollytics-go/utils"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

var (
	ErrUserUnauthorized = echo.NewHTTPError(http.StatusUnauthorized, "user unauthorized")
)

func Authenticate(h handlers.Handlers) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Writer.Header().Add("Vary", "Authorization")

			authorizationHeader := c.Request().Header.Get("Authorization")
			if authorizationHeader == "" {
				err := fmt.Errorf("authorization header not found")
				h.Utils.UserUnAuthorizedResponse(c, err)
				return ErrUserUnauthorized
			}

			headerParts := strings.Split(authorizationHeader, " ")
			if len(headerParts) != 2 || headerParts[0] != "Bearer" {
				err := fmt.Errorf("invalid authorization header")
				h.Utils.UserUnAuthorizedResponse(c, err)
				return ErrUserUnauthorized
			}

			token := headerParts[1]

			claims, err := jwt.HMACCheck([]byte(token), []byte(h.Config.JWT.Secret))
			if err != nil {
				h.Utils.UserUnAuthorizedResponse(c, err)
				return ErrUserUnauthorized
			}

			if !claims.Valid(time.Now()) {
				h.Utils.CustomErrorResponse(c, utils.Cake{"token expired": "Send refresh token"}, http.StatusUnauthorized, ErrUserUnauthorized)
				return ErrUserUnauthorized
			}

			if claims.Issuer != h.Config.JWT.Issuer {
				err := fmt.Errorf("invalid issuer")
				h.Utils.UserUnAuthorizedResponse(c, err)
				return ErrUserUnauthorized
			}

			reddit_uid := claims.Subject

			c.Set("reddit_uid", reddit_uid)

			return next(c)
		}
	}
}

func OptionalAuthenticate(h handlers.Handlers) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Writer.Header().Add("Vary", "Authorization")

			authorizationHeader := c.Request().Header.Get("Authorization")
			if authorizationHeader == "" {
				c.Set("reddit_uid", "")
				return next(c)
			}

			headerParts := strings.Split(authorizationHeader, " ")
			if len(headerParts) != 2 || headerParts[0] != "Bearer" {
				err := fmt.Errorf("invalid authorization header")
				h.Utils.UserUnAuthorizedResponse(c, err)
				return ErrUserUnauthorized
			}

			token := headerParts[1]

			claims, err := jwt.HMACCheck([]byte(token), []byte(h.Config.JWT.Secret))
			if err != nil {
				h.Utils.UserUnAuthorizedResponse(c, err)
				fmt.Println("Error in hmac check")
				return err
			}

			if !claims.Valid(time.Now()) {
				h.Utils.CustomErrorResponse(c, utils.Cake{"token expired": "Send refresh token"}, http.StatusUnauthorized, ErrUserUnauthorized)
				fmt.Println("Error in time check")
				return ErrUserUnauthorized
			}

			if claims.Issuer != h.Config.JWT.Issuer {
				err := fmt.Errorf("invalid issuer")
				h.Utils.UserUnAuthorizedResponse(c, err)
				return err
			}

			reddit_uid := claims.Subject

			c.Set("reddit_uid", reddit_uid)

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
				ip := realip.FromRequest(c.Request())

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
