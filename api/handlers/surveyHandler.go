package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/priyankishorems/bollytics-go/internal/data"
	"github.com/priyankishorems/bollytics-go/utils"
)

func (h *Handlers) CreateSurveyHandler(c echo.Context) error {
	// todo Uncomment before deploying
	// reddit_uid := c.Get("reddit_id").(string)
	reddit_uid := reddit_uid_test

	var input struct {
		Subreddit   string    `json:"subreddit" validate:"required"`
		Title       string    `json:"title" validate:"required"`
		Description string    `json:"description"`
		EndTime     time.Time `json:"end_time"`
	}

	if err := h.Utils.ReadJSON(c, &input); err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("error in reading json; %v", err))
		return err
	}

	if input.EndTime.IsZero() {
		input.EndTime = time.Now().Add(24 * time.Hour)
	}

	if err := h.Validate.Struct(input); err != nil {
		h.Utils.ValidationError(c, err)
		return err
	}

	id, err := h.Data.Surveys.CreateSurvey(reddit_uid, input.Subreddit, input.Title, input.Description, input.EndTime)
	if err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("Error creating survey: %v", err))
		return err
	}

	return c.JSON(http.StatusCreated, Cake{"id": id})
}

func (h *Handlers) CreateSurveyQuestionsHandler(c echo.Context) error {

	// todo Uncomment before deploying
	// reddit_uid := c.Get("reddit_id").(string)
	reddit_uid := reddit_uid_test
	SurveyID, err := h.Utils.ReadIntParam(c, "survey_id")
	if err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("Invalid Survey ID: %v", err))
		return err
	}

	owner, err := h.Data.Surveys.GetSurveyOwner(reddit_uid, SurveyID)
	if err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("error in getting survey owner; %v", err))
		return err
	}

	if !owner {
		h.Utils.UserUnAuthorizedResponse(c)
		return nil
	}

	questions := new([]data.Question)

	if err := h.Utils.ReadJSON(c, &questions); err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("error in reading json; %v", err))
		return err
	}

	var allOptions [][]data.SuveyQuestionOption

	for i := range *questions {
		question := &(*questions)[i]

		var options []data.SuveyQuestionOption

		if err := json.Unmarshal(question.Options, &options); err != nil {
			h.Utils.InternalServerError(c, fmt.Errorf("error in unmarshalling options; %v", err))
			return err
		}

		switch question.QuestionType {
		case "multiple_choice", "single_choice":
			if len(options) < 2 {
				h.Utils.BadRequest(c, fmt.Errorf("%s question should have at least 2 options", question.QuestionType))
				return err
			}
		case "text":
			if len(options) > 0 {
				h.Utils.BadRequest(c, fmt.Errorf("Text question should not have any options"))
				return err
			}
		case "":
			question.QuestionType = "single_choice"
		}

		if err := h.Validate.Struct(question); err != nil {
			h.Utils.ValidationError(c, err)
			return err
		}

		for _, option := range options {
			if err := h.Validate.Struct(option); err != nil {
				h.Utils.ValidationError(c, err)
				return err
			}
		}
		allOptions = append(allOptions, options)
	}

	if err := h.Data.Surveys.CreateSurveyQuestions(SurveyID, questions, allOptions); err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("Error creating survey questions: %v", err))
		return err
	}

	return c.JSON(http.StatusCreated, Cake{"message": "Survey Questions Created Successfully"})
}

type response_data struct {
	ResponseData json.RawMessage `json:"response_data" validate:"required"`
}

func (h *Handlers) CreateSurveyResponseHandler(c echo.Context) error {
	// todo Uncomment before deploying
	// reddit_uid := c.Get("reddit_id").(string)
	reddit_uid := reddit_uid_test
	SurveyID, err := h.Utils.ReadIntParam(c, "survey_id")
	if err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("Invalid Survey ID: %v", err))
		return err
	}

	var input *response_data

	if err = h.Utils.ReadJSON(c, &input); err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("error in reading json; %v", err))
		return err
	}

	var allResponses []data.QuestionResponse

	if err := json.Unmarshal(input.ResponseData, &allResponses); err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("error in unmarshalling response data; %v", err))
		return err
	}

	if err := h.Validate.Struct(input); err != nil {
		h.Utils.ValidationError(c, err)
		return err
	}

	for _, response := range allResponses {
		if response.ID == 0 {
			h.Utils.CustomErrorResponse(c, utils.Cake{"error": "response data should have question_id"}, http.StatusBadRequest, nil)
			return err
		}
		if err := h.Validate.Struct(response); err != nil {
			h.Utils.ValidationError(c, err)
			return err
		}
	}

	if err := h.Data.Surveys.CreateSurveyResponse(SurveyID, reddit_uid, allResponses); err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("Error creating survey response: %v", err))
		return err
	}

	return c.JSON(http.StatusCreated, Cake{"message": "Survey Response Created Successfully"})
}

func (h *Handlers) GetSurveyByIDHandler(c echo.Context) error {
	SurveyID, err := h.Utils.ReadIntParam(c, "survey_id")
	if err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("Invalid Survey ID: %v", err))
		return err
	}

	survey, err := h.Data.Surveys.GetSurveyQuestionByID(SurveyID)
	if err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("Error getting survey: %v", err))
		return err
	}

	return c.JSON(http.StatusOK, survey)
}

// todo should be let if the requester is creator of the survey or if the survey is public
func (h *Handlers) GetSurveyResponsesByIDHandler(c echo.Context) error {
	SurveyID, err := h.Utils.ReadIntParam(c, "survey_id")
	if err != nil {
		h.Utils.BadRequest(c, fmt.Errorf("Invalid Survey ID: %v", err))
		return err
	}

	responses, err := h.Data.Surveys.GetSurveyResponses(SurveyID)
	if err != nil {
		h.Utils.InternalServerError(c, fmt.Errorf("Error getting survey responses: %v", err))
		return err
	}

	return c.JSON(http.StatusOK, responses)
}
