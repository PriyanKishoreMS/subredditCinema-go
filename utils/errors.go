package utils

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type cake map[string]interface{}

func (u *utilsImpl) logError(c echo.Context, err error) {
	log.Error(err)
}

func (u *utilsImpl) resposeError(c echo.Context, status int, message interface{}) {
	err := c.JSON(status, cake{"error": message})
	if err != nil {
		u.logError(c, err)
		c.Response().WriteHeader(http.StatusInternalServerError)
	}
}

func (u *utilsImpl) InternalServerError(c echo.Context, err error) {
	u.logError(c, err)
	message := "server encountered an error and could not process your request"
	u.resposeError(c, http.StatusInternalServerError, message)
}

func (u *utilsImpl) BadRequest(c echo.Context, err error) {
	u.logError(c, err)
	message := "you have given a bad or invalid request, please try again"
	u.resposeError(c, http.StatusBadRequest, message)
}

func (u *utilsImpl) MethodNotFound(c echo.Context) {
	message := "the method is not allowed"
	u.resposeError(c, http.StatusMethodNotAllowed, message)
}

func (u *utilsImpl) NotFoundResponse(c echo.Context) {
	message := "the request is not found"
	u.resposeError(c, http.StatusNotFound, message)
}

func (u *utilsImpl) EditConflictResponse(c echo.Context) {
	message := "unable to update the record due to edit conflict, please try again"
	u.resposeError(c, http.StatusConflict, message)
}

func (u *utilsImpl) UserUnAuthorizedResponse(c echo.Context) {
	message := "You are not authorized to access this"
	u.resposeError(c, http.StatusUnauthorized, message)
}

func (u *utilsImpl) RateLimitExceededResponse(c echo.Context) {
	message := "Rate limit exceeded"
	u.resposeError(c, http.StatusTooManyRequests, message)
}

func (u *utilsImpl) CustomErrorResponse(c echo.Context, message cake, status int, err error) {
	u.logError(c, err)
	u.resposeError(c, status, message)
}

func (u *utilsImpl) ValidationError(c echo.Context, err error) {
	validationError := make(map[string]interface{})
	validErrs := err.(validator.ValidationErrors)
	for _, e := range validErrs {
		var errMsg string

		switch e.Tag() {
		case "required":
			errMsg = "is required"
		case "email":
			errMsg = fmt.Sprint(e.Field(), " must be a type of email")
		case "gte":
			errMsg = "value must be greater than 0"
		case "lte":
			errMsg = "value must be lesser than the given value"

		default:
			errMsg = fmt.Sprintf("Validation error on %s: %s", e.Field(), e.Tag())
		}

		validationError[e.Field()] = errMsg
	}
	u.resposeError(c, http.StatusUnprocessableEntity, validationError)
}
