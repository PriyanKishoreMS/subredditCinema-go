package data

const (
	CreateSurveyQuery = `
	INSERT INTO surveys (reddit_uid, subreddit, title, description, end_time)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id;
	`

	CreateSurveyQuestionQuery = `
	INSERT INTO survey_questions (survey_id, question_order, question_text, question_type, is_required)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id;
	`

	CreateSurveyOptionQuery = `
	INSERT INTO survey_options (question_id, option_order, option_text)
	VALUES ($1, $2, $3);
	`

	CreateResponseQuery = `
	INSERT INTO survey_responses (survey_id, reddit_uid)
	VALUES ($1, $2)
	RETURNING id;
	`

	CreateAnswerQuery = `
	INSERT INTO survey_answers (response_id, question_id, answer_text, selected_option_id)
	VALUES ($1, $2, $3, $4);
	`

	GetQuestionTypeQuery = `
	SELECT id, question_type FROM survey_questions WHERE survey_id = $1
	`

	GetAllSurveyDetailsQuery = `
	SELECT 
    	COUNT(*) OVER () AS total,
    	s.id,
    	s.subreddit,
    	s.title,
    	s.description,
    	s.end_time,
    	s.is_result_public,
		s.created_at,
    	u.username,
    	u.avatar,
		u.reddit_uid,
    	COALESCE(sr.response_count, 0) AS total_responses
	FROM
    	surveys s
	INNER JOIN
    	users u ON s.reddit_uid = u.reddit_uid
	LEFT JOIN
    	(SELECT survey_id, COUNT(*) AS response_count 
     	FROM survey_responses 
     	GROUP BY survey_id) sr ON sr.survey_id = s.id
	ORDER BY
    	s.created_at DESC
	LIMIT $1
	OFFSET $2
	`

	GetSubSurveyDetailsQuery = `
	SELECT 
    	COUNT(*) OVER () AS total,
    	s.id,
    	s.subreddit,
    	s.title,
    	s.description,
    	s.end_time,
    	s.is_result_public,
		s.created_at,
    	u.username,
    	u.avatar,
    	COALESCE(sr.response_count, 0) AS total_responses
	FROM
    	surveys s
	INNER JOIN
    	users u ON s.reddit_uid = u.reddit_uid
	LEFT JOIN
    	(SELECT survey_id, COUNT(*) AS response_count
     	FROM survey_responses
     	GROUP BY survey_id) sr ON sr.survey_id = s.id
	where s.subreddit = $1
	ORDER BY
    	s.created_at DESC
	LIMIT $2
	OFFSET $3
	`

	GetSurveyDetailsQuery = `
	SELECT 
    s.id,
    s.subreddit,
    s.title,
    s.description,
    s.end_time,
    s.is_result_public,
    u.username,
    u.avatar
	FROM 
    	surveys s
	inner join users u on s.reddit_uid = u.reddit_uid
	where s.id = $1;
	`

	GetSurveyQuestionsQuery = `
	SELECT 
    	id AS question_id,
    	question_order,
    	question_text,
    	question_type,
    	is_required
	FROM 
    	survey_questions
	WHERE 
    	survey_id = $1
	ORDER BY 
    	question_order;
	`

	GetSurveyOptionsQuery = `
	SELECT 
    so.id AS option_id,
    so.question_id,
    so.option_order,
    so.option_text
	FROM 
    	survey_options so
	JOIN 
    	survey_questions sq ON so.question_id = sq.id
	WHERE 
    	sq.survey_id = $1
	ORDER BY 
    	sq.question_order, so.option_order;
	`

	GetSurveyAnswerCountsQuery = `
	SELECT 
    sq.id AS question_id,
    so.id AS option_id,
    COUNT(sa.selected_option_id) AS selection_count
	FROM 
    	survey_questions sq
	JOIN 
    	survey_options so ON sq.id = so.question_id
	LEFT JOIN 
    	survey_answers sa ON so.id = sa.selected_option_id
	WHERE 
    	sq.survey_id = $1
	GROUP BY 
    	sq.id, sq.question_text, so.id, so.option_text
	ORDER BY 
    	sq.id, so.id;
	`

	GetResponsesToEachQuestionQuery = `
	SELECT 
    sq.id AS question_id,
    COUNT(DISTINCT sa.response_id) AS response_count
	FROM 
    	survey_questions sq
	LEFT JOIN 
    	survey_answers sa ON sq.id = sa.question_id
	WHERE 
    	sq.survey_id = $1 and sq.question_type != 'text'
	GROUP BY 
    	sq.id
	ORDER BY 
    	sq.id;
	`
)
