package handlers

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	Validate validator.Validate
}

func (h *Handlers) HomeFunc(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "start",
	})
}
