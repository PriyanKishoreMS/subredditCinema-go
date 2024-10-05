package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pascaldekloe/jwt"
	"github.com/priyankishorems/bollytics-go/internal/data"
	"github.com/priyankishorems/bollytics-go/utils"
	"golang.org/x/oauth2"
)

type AuthResponse struct {
	Username string `json:"username"`
}

var (
	ErrUserUnauthorized = echo.NewHTTPError(http.StatusUnauthorized, "user unauthorized")
)

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
		h.Utils.BadRequest(c, fmt.Errorf("failed to exchange token: %s", err))
		return err
	}

	userdata, err := h.GetAuthUserDataFromReddit(c, token, utils.RedditUserAgentWeb)
	if err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("failed to get user: %s", err))
		return err
	}

	user, err := h.Data.Users.InsertUser(userdata.Name, userdata.Avatar, userdata.RedditID)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return err
	}

	accessToken, RefreshToken, err := data.GenerateAuthTokens(userdata.RedditID, h.Config.JWT.Secret, h.Config.JWT.Issuer)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return err
	}

	data := Cake{
		"accessToken":  string(accessToken),
		"refreshToken": string(RefreshToken),
		"user":         user,
	}

	tokensJSON, err := json.Marshal(data)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return err
	}

	fmt.Println("tokensJSON here", string(tokensJSON))

	c.SetCookie(&http.Cookie{
		Name:     "tokens",
		Value:    url.QueryEscape(string(tokensJSON)),
		Path:     "/",
		MaxAge:   60 * 60,
		Domain:   ".priyankishore.dev",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	return c.Redirect(http.StatusFound, "https://subcinema.priyankishore.dev/auth-callback")
}

func (h *Handlers) RefreshTokenHandler(c echo.Context) error {

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

	id := claims.Subject
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return err
	}

	accessToken, err := data.GenerateAccessToken(id, []byte(h.Config.JWT.Secret), h.Config.JWT.Issuer)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return err
	}
	return c.JSON(200, Cake{"accessToken": string(accessToken)})
}
