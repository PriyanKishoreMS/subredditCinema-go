package data

const (
	CreateSurveyQuery = `
	INSERT INTO surveys (reddit_uid, subreddit, title, description, end_time)
	VALUES ($1, $2, $3, $4, $5)
	`

	CreateSurveyQuestionQuery = `
	INSERT INTO survey_questions (survey_id, question_order, question_text, question_type, options, is_required)
	VALUES ($1, $2, $3, $4, $5, $6) 
	`

	CreateSurveyResponseQuery = `
	INSERT INTO survey_responses (survey_id, reddit_uid, response_data) VALUES ($1, $2, $3)
	`

	GetSurveyOwnerQuery = `
	SELECT EXISTS(SELECT 1 FROM surveys WHERE reddit_uid = $1 AND id = $2)	
	`
)
