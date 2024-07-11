package handlers

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/priyankishorems/bollytics-go/internal/data"
	"github.com/priyankishorems/bollytics-go/utils"
)

type cake map[string]interface{}
type Handlers struct {
	Config   utils.Config
	Validate validator.Validate
	Utils    utils.Utilities
	Data     data.Models
}

func (h *Handlers) HomeFunc(c echo.Context) error {
	return c.JSON(http.StatusOK, cake{"message": "Welcome to Bollytics"})
}
