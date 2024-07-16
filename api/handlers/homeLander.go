package handlers

import (
	"net/http"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/priyankishorems/bollytics-go/internal/data"
	"github.com/priyankishorems/bollytics-go/utils"
	graw "github.com/turnage/graw/reddit"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

type Cake map[string]interface{}
type Handlers struct {
	Config    utils.Config
	Validate  validator.Validate
	Utils     utils.Utilities
	Data      data.Models
	Tmdb      *tmdb.Client
	RedditBot graw.Bot
	Reddit    *reddit.Client
}

func (h *Handlers) HomeFunc(c echo.Context) error {
	data := Cake{
		"message": "Welcome to Bollytics API",
		"status":  "available",
		"system_info": Cake{
			"environment": h.Config.Env,
			"port":        h.Config.Port,
		},
	}
	return c.JSON(http.StatusOK, data)
}
