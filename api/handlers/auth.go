package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/priyankishorems/bollytics-go/utils"
	"golang.org/x/oauth2"
)

type AuthResponse struct {
	Username string `json:"username"`
}

func generateRandomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (h *Handlers) LoginHandler(c echo.Context) error {
	state := generateRandomState()
	url := utils.OauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *Handlers) CallbackHandler(c echo.Context) error {
	code := h.Utils.ReadStringQuery(c.QueryParams(), "code", "")
	if code == "" {
		h.Utils.BadRequest(c, fmt.Errorf("no code provided"))
		return fmt.Errorf("no code provided")
	}

	token, err := utils.OauthConfig.Exchange(c.Request().Context(), code)
	if err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("Failed to exchange token: %s", err))
		return err
	}

	user, err := h.GetAuthUserDataFromReddit(c, token, utils.RedditUserAgentWeb)
	if err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("Failed to get user: %s", err))
		return err
	}

	return c.JSON(http.StatusOK, user)
}
