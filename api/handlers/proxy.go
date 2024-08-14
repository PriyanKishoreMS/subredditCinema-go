package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) ProxyHandler(c echo.Context) error {
	imageURL, err := h.Utils.ReadStringParam(c, "url")
	if err != nil {
		h.Utils.BadRequest(c, err)
		return err
	}

	resp, err := http.Get(imageURL)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return err
	}

	defer resp.Body.Close()

	return c.Stream(http.StatusOK, resp.Header.Get("Content-Type"), resp.Body)
}
