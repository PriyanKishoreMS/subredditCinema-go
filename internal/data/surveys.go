package data

import (
	"encoding/json"
	"time"

	pgx "github.com/jackc/pgx/v5/pgxpool"
)

type SurveysModel struct {
	DB *pgx.Pool
}

type SuveyQuestionOption struct {
	ID   int    `json:"id" validate:"required"`
	Text string `json:"text" validate:"required"`
}

type Question struct {
	QuestionText string          `json:"question_text" validate:"required"`
	QuestionType string          `json:"question_type"`
	Options      json.RawMessage `json:"options"`
	IsRequired   bool            `json:"is_required"`
}

type QuestionResponse struct {
	ID              int    `json:"id" validate:"required"`
	Text            string `json:"text,omitempty"`
	OptionID        int    `json:"option_id,omitempty"`
	MultipleOptions []int  `json:"multiple_options_id,omitempty"`
}

func (s SurveysModel) CreateSurvey(redditUID, subreddit, title, description string, endTime time.Time) error {
	ctx, cancel := Handlectx()
	defer cancel()

	query := CreateSurveyQuery

	_, err := s.DB.Exec(ctx, query, redditUID, subreddit, title, description, endTime)
	if err != nil {
		return err
	}

	return nil
}

func (s SurveysModel) CreateSurveyQuestions(surveyID int, questions *[]Question, options [][]SuveyQuestionOption) error {
	ctx, cancel := Handlectx()
	defer cancel()

	query := CreateSurveyQuestionQuery

	for i, question := range *questions {
		_, err := s.DB.Exec(ctx, query, surveyID, i+1, question.QuestionText, question.QuestionType, options[i], question.IsRequired)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s SurveysModel) CreateSurveyResponse(surveyID int, redditUID string, response_data []QuestionResponse) error {
	ctx, cancel := Handlectx()
	defer cancel()

	query := CreateSurveyResponseQuery

	_, err := s.DB.Exec(ctx, query, surveyID, redditUID, response_data)
	if err != nil {
		return err
	}

	return nil
}

func (s SurveysModel) GetSurveyOwner(reddit_uid string, survey_id int) (bool, error) {
	ctx, cancel := Handlectx()
	defer cancel()

	query := GetSurveyOwnerQuery

	var owner bool
	err := s.DB.QueryRow(ctx, query, reddit_uid, survey_id).Scan(&owner)
	if err != nil {
		return false, err
	}

	return owner, nil
}
