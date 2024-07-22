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

type FullQuestionData struct {
	SurveyID    int             `json:"survey_id"`
	RedditUID   string          `json:"reddit_uid"`
	Subreddit   string          `json:"subreddit"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	StartTime   time.Time       `json:"start_time"`
	EntTime     time.Time       `json:"end_time"`
	IsActive    bool            `json:"is_active"`
	Questions   json.RawMessage `json:"questions"`
}

type QuestionResponse struct {
	ID int `json:"id" validate:"required"`
	// Text            string `json:"text,omitempty"`
	OptionID int `json:"option_id" validate:"required"`
	// MultipleOptions []int  `json:"multiple_options_id,omitempty"`
}

type FullResponseData struct {
	QuestionID             int             `json:"question_id"`
	OptionCounts           json.RawMessage `json:"option_counts"`
	TotalQuestionResponses int             `json:"total_question_responses"`
}

func (s SurveysModel) CreateSurvey(redditUID, subreddit, title, description string, endTime time.Time) (int, error) {
	ctx, cancel := Handlectx()
	defer cancel()

	query := CreateSurveyQuery

	var id int
	row := s.DB.QueryRow(ctx, query, redditUID, subreddit, title, description, endTime)
	err := row.Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
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

func (s SurveysModel) GetSurveyQuestionByID(surveyID int) (*FullQuestionData, error) {

	ctx, cancel := Handlectx()
	defer cancel()

	query := GetSurveyQuestionsQuery

	var survey FullQuestionData
	err := s.DB.QueryRow(ctx, query, surveyID).Scan(&survey.SurveyID, &survey.RedditUID, &survey.Subreddit, &survey.Title, &survey.Description, &survey.StartTime, &survey.EntTime, &survey.IsActive, &survey.Questions)
	if err != nil {
		return nil, err
	}

	return &survey, nil
}

func (s SurveysModel) GetSurveyResponses(surveyID int) ([]FullResponseData, error) {
	ctx, cancel := Handlectx()
	defer cancel()

	query := GetSurveyResponseDataQuery

	rows, err := s.DB.Query(ctx, query, surveyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var responses []FullResponseData
	for rows.Next() {
		var response FullResponseData
		err := rows.Scan(&response.QuestionID, &response.OptionCounts, &response.TotalQuestionResponses)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}

	return responses, nil
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
