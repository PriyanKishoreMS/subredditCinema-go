package data

import (
	"fmt"
	"time"

	pgx "github.com/jackc/pgx/v5/pgxpool"
)

type SurveysModel struct {
	DB *pgx.Pool
}

type Survey struct {
	SurveyID       int        `json:"survey_id"`
	Username       string     `json:"username"`
	Avatar         string     `json:"avatar"`
	RedditUID      string     `json:"reddit_uid,omitempty"`
	Subreddit      string     `json:"subreddit" validate:"required"`
	Title          string     `json:"title" validate:"required"`
	Description    string     `json:"description"`
	StartTime      time.Time  `json:"start_time" validate:"required"`
	EndTime        time.Time  `json:"end_time" validate:"required"`
	IsResultPublic bool       `json:"is_result_public"`
	TotalResponses int        `json:"total_responses"`
	Questions      []Question `json:"questions,omitempty" validate:"required,dive"`
}

type Question struct {
	QuestionID int      `json:"question_id"`
	Order      int      `json:"order" validate:"required"`
	Text       string   `json:"text" validate:"required"`
	Type       string   `json:"type" validate:"required"`
	IsRequired bool     `json:"is_required"`
	Options    []Option `json:"options" validate:"dive"`
}

type Option struct {
	QuestionID int    `json:"question_id"`
	OptionID   int    `json:"option_id"`
	Order      int    `json:"order" validate:"required"`
	Text       string `json:"text" validate:"required"`
}

type Answers struct {
	QuestionID       int    `json:"question_id" validate:"required"`
	AnswerText       string `json:"answer_text"`
	SelectedOptionID *int   `json:"selected_option_id"`
}

func (s SurveysModel) CreateSurvey(survey *Survey) (err error) {
	ctx, cancel := Handlectx()
	defer cancel()

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		err = fmt.Errorf("error in starting transaction; %v", err)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(ctx)
			err = fmt.Errorf("transaction panicked: %v", r)
		} else if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()
	query := CreateSurveyQuery
	var surveyID int

	err = tx.QueryRow(ctx, query, survey.RedditUID, survey.Subreddit, survey.Title, survey.Description, survey.StartTime, survey.EndTime).Scan(&surveyID)
	if err != nil {
		return
	}

	for _, q := range survey.Questions {
		var quesionID int
		err = tx.QueryRow(ctx, CreateSurveyQuestionQuery, surveyID, q.Order, q.Text, q.Type, q.IsRequired).Scan(&quesionID)
		if err != nil {
			return
		}

		for _, o := range q.Options {
			_, err = tx.Exec(ctx, CreateSurveyOptionQuery, quesionID, o.Order, o.Text)
			if err != nil {
				return
			}
		}
	}
	return nil
}

func (s SurveysModel) CreateSurveyResponses(redditUID string, surveyID int, answers *[]Answers) (err error) {
	ctx, cancel := Handlectx()
	defer cancel()

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error in starting transaction; %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(ctx)
			err = fmt.Errorf("transaction panicked: %v", r)
		} else if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	questionTypes := make(map[int]string)
	rows, err := tx.Query(ctx, GetQuestionTypeQuery, surveyID)
	if err != nil {
		err = fmt.Errorf("error fetching question types: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var qType string
		if err = rows.Scan(&id, &qType); err != nil {
			err = fmt.Errorf("error scanning question types: %v", err)
			return
		}
		questionTypes[id] = qType
	}

	query := CreateResponseQuery
	var responseID int

	err = tx.QueryRow(ctx, query, surveyID, redditUID).Scan(&responseID)
	if err != nil {
		return
	}

	for _, a := range *answers {
		questionType, exists := questionTypes[a.QuestionID]
		if !exists {
			err = fmt.Errorf("question %d not found in survey", a.QuestionID)
			return
		}

		if err = validateAnswer(questionType, a); err != nil {
			return
		}

		_, err = tx.Exec(ctx, CreateAnswerQuery, responseID, a.QuestionID, a.AnswerText, a.SelectedOptionID)
		if err != nil {
			return
		}
	}

	return nil
}

func validateAnswer(questionType string, answer Answers) error {
	switch questionType {
	case "text":
		if answer.AnswerText == "" {
			return fmt.Errorf("text answer required for question %d", answer.QuestionID)
		}
	case "multiple":
		if answer.SelectedOptionID == nil {
			return fmt.Errorf("option selection required for question %d", answer.QuestionID)
		}
	case "single":
		if answer.SelectedOptionID == nil {
			return fmt.Errorf("option selection required for question %d", answer.QuestionID)
		}
	default:
		return fmt.Errorf("unknown question type: %s", questionType)
	}
	return nil
}

func (s SurveysModel) GetSurveyByID(surveyID int) (*Survey, error) {
	ctx, cancel := Handlectx()
	defer cancel()

	survey := new(Survey)

	if err := s.DB.QueryRow(ctx, GetSurveyDetailsQuery, surveyID).Scan(&survey.SurveyID, &survey.Subreddit, &survey.Title, &survey.Description, &survey.StartTime, &survey.EndTime, &survey.IsResultPublic, &survey.Username, &survey.Avatar); err != nil {
		return nil, err
	}

	rows, err := s.DB.Query(ctx, GetSurveyQuestionsQuery, surveyID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		q := new(Question)
		if err := rows.Scan(&q.QuestionID, &q.Order, &q.Text, &q.Type, &q.IsRequired); err != nil {
			return nil, err
		}
		survey.Questions = append(survey.Questions, *q)
	}

	rows, err = s.DB.Query(ctx, GetSurveyOptionsQuery, surveyID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		o := new(Option)
		if err := rows.Scan(&o.OptionID, &o.QuestionID, &o.Order, &o.Text); err != nil {
			return nil, err
		}
		for i, q := range survey.Questions {
			if q.QuestionID == o.QuestionID {
				survey.Questions[i].Options = append(survey.Questions[i].Options, *o)
			}
		}
	}

	return survey, nil
}

func (s SurveysModel) GetAllSurveys(filters Filters) ([]Survey, Metadata, error) {
	ctx, cancel := Handlectx()
	defer cancel()

	query := GetAllSurveyDetailsQuery

	var surveys []Survey

	rows, err := s.DB.Query(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	for rows.Next() {
		survey := new(Survey)
		if err := rows.Scan(&totalRecords, &survey.SurveyID, &survey.Subreddit, &survey.Title, &survey.Description, &survey.StartTime, &survey.EndTime, &survey.IsResultPublic, &survey.Username, &survey.Avatar, &survey.TotalResponses); err != nil {
			return nil, Metadata{}, err
		}
		surveys = append(surveys, *survey)
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return surveys, metadata, nil
}
