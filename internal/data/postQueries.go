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
	select extract(
        	hour
        	from created_utc
    	) as hour,
    	extract(
        	dow
        	from created_utc
    	) as day,
    	count(*) as post_count
	from subreddit_posts
	where subreddit = $1
    	and created_utc >= now() - make_interval(days := $2)
	group by subreddit,
    	hour,
    	day
	order by day asc;
	`
)
