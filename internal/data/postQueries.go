package data

const (
	InsertPostsQuery = `
		INSERT INTO subreddit_posts (
    	id,
    	name,
    	created_utc,
    	permalink,
    	title,
    	category,
    	selftext,
    	score,
    	upvote_ratio,
    	num_comments,
    	subreddit,
    	subreddit_id,
    	subreddit_subscribers,
    	author,
    	author_fullname
	)
	VALUES (
    	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
	)
	ON CONFLICT(id) DO
	UPDATE
	SET 
    	score = EXCLUDED.score,
    	upvote_ratio = EXCLUDED.upvote_ratio,
    	num_comments = EXCLUDED.num_comments,
		version = subreddit_posts.version + 1,
    	top_and_controversial = CASE
        	WHEN subreddit_posts.category <> EXCLUDED.category
        	OR (subreddit_posts.category = 'controversial' AND EXCLUDED.category = 'top')
        	THEN TRUE
        	ELSE subreddit_posts.top_and_controversial
    	END
	`

	DeleteOldPostsQuery = `
	DELETE FROM subreddit_posts
	WHERE created_utc < NOW() - INTERVAL '365 days'
	`

	TopUsersQuery = `
	select author as user,
    	count(*) as author_count
	from subreddit_posts
	where subreddit = $1
    	and (
        	category = 'top'
        	or (
            	category = 'controversial'
            	and top_and_controversial = true
        	)
    	)
    	and created_utc > now() - make_interval(days := $2)
		and author != '[deleted]'
	group by author
	order by author_count desc
	limit 5
	`

	ControversialUsersQuery = `
	select author as user,
    	count(*) as author_count
	from subreddit_posts
	where subreddit = $1
    	and (
        	category = 'controversial'
        	or (
            	category = 'top'
            	and top_and_controversial = true
        	)
    	)
    	and created_utc > now() - make_interval(days := $2)
		and author != '[deleted]'
	group by author
	order by author_count desc
	limit 5
	`

	TopPostsQuery = `
	select id,
    	title,
    	selftext,
    	author,
    	permalink,
    	score,
    	upvote_ratio,
		subreddit,
    	num_comments,
    	category,
    	round((score * upvote_ratio)::numeric, 2) as top_score
	from subreddit_posts
	where subreddit = $1
    	and category = 'top'
    	and created_utc > now() - make_interval(days := $2)
    	and top_and_controversial = false
	order by top_score desc
	limit 5
	`

	ControversialPostsQuery = `
	select id,
    	title,
    	selftext,
    	author,
    	permalink,
    	score,
    	upvote_ratio,
		subreddit,
    	num_comments,
    	category,
    	round(
        	(score * (1 - upvote_ratio) * num_comments)::numeric,
        	2
    	) as controversary_score
	from subreddit_posts
	where subreddit = $1
    	and category = 'controversial'
    	and created_utc > now() - make_interval(days := $2)
    	and top_and_controversial = false
	order by controversary_score desc
	limit 5
	`

	MostHatedPostsQuery = `
	select id,
    	title,
    	selftext,
    	author,
    	permalink,
    	score,
    	upvote_ratio,
		subreddit,
    	num_comments,
    	category,
		upvote_ratio as category_score
	from subreddit_posts
	where subreddit = $1
    	and category = 'controversial'
    	and created_utc > now() - make_interval(days := $2)
    	and top_and_controversial = false
	order by category_score asc
	limit 5
	`

	TopAndControversialPostsQuery = `
	select id,
    	title,
    	selftext,
    	author,
    	permalink,
    	score,
    	upvote_ratio,
		subreddit,
    	num_comments,
    	category,
    	round((score * upvote_ratio)::numeric, 2) as top_score
	from subreddit_posts
	where subreddit = $1
    	and top_and_controversial = true
    	and created_utc > now() - make_interval(days := $2)
	order by top_score desc
	limit 5
	`

	FrequencyOfPostsQuery = `
	WITH date_range AS (
    SELECT 
        (date_trunc('week', CURRENT_DATE - INTERVAL '4 weeks')::date + make_interval(days := $2)) AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Kolkata' AS start_date,
        (date_trunc('week', CURRENT_DATE)::date + make_interval(days := $2)) AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Kolkata' AS end_date
	)
	SELECT 
    	EXTRACT(HOUR FROM (created_utc AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Kolkata')) AS hour,
    	EXTRACT(DOW FROM (created_utc AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Kolkata')) AS day,
    	COUNT(*) AS post_count 
	FROM subreddit_posts, date_range
	WHERE 
    	created_utc >= date_range.start_date
    	AND created_utc < date_range.end_date
    	AND subreddit = $1
	GROUP BY 
    	hour, day 
	ORDER BY 
    	day ASC, hour ASC;
	`

	GetAllTextsOfInterval = `
    SELECT 
      	title || ' ' || selftext AS full_text 
    FROM 
      	subreddit_posts 
    WHERE 
      	subreddit = $1 
      	AND created_utc >= now() - make_interval(days := $2)
	`

	InsertUserQuery = `	
    INSERT INTO users (reddit_uid, username, avatar) 
    VALUES 
      	($1, $2, $3) ON CONFLICT (username) DO 
    UPDATE 
    SET 
      	avatar = $3, 
      	version = users.version + 1,
		last_login = NOW()
	RETURNING id, reddit_uid, username, avatar
	`

	CheckUserExistsQuery = `
	SELECT EXISTS(
		SELECT 1
		FROM users
		WHERE reddit_uid = $1
	)
	`
)
