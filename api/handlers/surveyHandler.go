package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/priyankishorems/bollytics-go/internal/data"
)

func (h *Handlers) CreateSurveyHandler(c echo.Context) error {
	// todo Uncomment before deploying
	// reddit_uid := c.Get("reddit_id").(string)
	reddit_uid := reddit_uid_test
	survey := new(data.Survey)

	survey.RedditUID = reddit_uid

	if survey.StartTime.IsZero() {
		survey.StartTime = time.Now()
	}

	if survey.EndTime.IsZero() {
		survey.EndTime = time.Now().Add(time.Hour * 24)
	}

	if err := h.Utils.ReadJSON(c, &survey); err != nil {
		h.Utils.BadRequest(c, err)
		return err
	}

	if err := h.Validate.Struct(survey); err != nil {
		h.Utils.ValidationError(c, err)
		return err
	}

	if err := h.Data.Surveys.CreateSurvey(survey); err != nil {
		h.Utils.InternalServerError(c, err)
		return err
	}

	return c.JSON(http.StatusAccepted, Cake{"success": "Survey Created"})
}

func (h *Handlers) CreateSurveyResponsesHandler(c echo.Context) error {
	// todo Uncomment before deploying
	// reddit_uid := c.Get("reddit_id").(string)
	reddit_uid := reddit_uid_test

	surveyID, err := h.Utils.ReadIntParam(c, "survey_id")
	if err != nil {
		h.Utils.BadRequest(c, err)
		return err
	}

	answers := new([]data.Answers)

	if err := h.Utils.ReadJSON(c, &answers); err != nil {
		h.Utils.BadRequest(c, err)
		return err
	}

	for _, answer := range *answers {
		if err := h.Validate.Struct(answer); err != nil {
			h.Utils.ValidationError(c, err)
			return err
		}
	}

	if err := h.Data.Surveys.CreateSurveyResponses(reddit_uid, surveyID, answers); err != nil {
		h.Utils.InternalServerError(c, err)
		return err
	}

	return c.JSON(http.StatusAccepted, Cake{"success": "Survey Responses Created"})
}

func (h *Handlers) GetSurveyByIDHandler(c echo.Context) error {
	surveyID, err := h.Utils.ReadIntParam(c, "survey_id")
	if err != nil {
		h.Utils.BadRequest(c, err)
		return err
	}

	survey, err := h.Data.Surveys.GetSurveyByID(surveyID)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return err
	}

	return c.JSON(http.StatusOK, survey)
}

func (h *Handlers) GetAllSurveysHandler(c echo.Context) error {
	filters := data.Filters{}

	qs := c.Request().URL.Query()
	filters.Page = h.Utils.ReadIntQuery(qs, "page", 1)
	filters.PageSize = h.Utils.ReadIntQuery(qs, "page_size", 10)

	err := h.Validate.Struct(filters)
	if err != nil {
		h.Utils.ValidationError(c, err)
		return err
	}

	surveys, metadata, err := h.Data.Surveys.GetAllSurveys(filters)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return err
	}

	return c.JSON(http.StatusOK, Cake{"surveys": surveys, "metadata": metadata})
}
