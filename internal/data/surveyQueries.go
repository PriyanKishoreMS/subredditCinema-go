package data

const (
	CreateSurveyQuery = `
	INSERT INTO surveys (reddit_uid, subreddit, title, description, end_time)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id
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

	GetSurveyQuestionsQuery = `
	SELECT 
    s.id AS survey_id, 
    s.reddit_uid, 
    s.subreddit, 
    s.title, 
    s.description, 
    s.start_time, 
    s.end_time, 
    s.is_active,
    json_agg(
        json_build_object(
            'id', sq.id,
            'question_order', sq.question_order,
            'question_text', sq.question_text,
            'question_type', sq.question_type,
            'options', sq.options,
            'is_required', sq.is_required
        ) ORDER BY sq.question_order
    ) AS questions
	FROM 
    	surveys s
	LEFT JOIN 
    	survey_questions sq ON s.id = sq.survey_id
	WHERE 
    	s.id = $1
	GROUP BY 
    	s.id
	`

	GetSurveyResponseDataQuery = `
	WITH response_counts AS (
    SELECT 
        survey_id,
        (jsonb_array_elements(response_data)->>'id')::int AS question_id,
        (jsonb_array_elements(response_data)->>'option_id')::int AS option_id
    FROM 
        survey_responses
    WHERE 
        survey_id = $1
	),
	all_options AS (
    	SELECT 
        	id AS question_id,
        	(jsonb_array_elements(options)->>'id')::int AS option_id
    	FROM 
        	survey_questions
    	WHERE 
        	survey_id = $1
	),
	option_counts AS (
    	SELECT 
        	ao.question_id,
        	ao.option_id,
        	COUNT(rc.option_id) AS count
    	FROM 
        	all_options ao
    	LEFT JOIN 
        	response_counts rc ON ao.question_id = rc.question_id AND ao.option_id = rc.option_id
    	GROUP BY 
        	ao.question_id, ao.option_id
	)
	SELECT 
    	question_id,
    	jsonb_object_agg(option_id, count ORDER BY option_id) AS option_counts,
    	SUM(count) AS total_question_responses
	FROM 
    	option_counts
	GROUP BY 
    	question_id
	ORDER BY 
    	question_id;
	`
)
