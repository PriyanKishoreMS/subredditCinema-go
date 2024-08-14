package data

const (
	CreatePollsQuery = `
	INSERT INTO polls (reddit_uid, subreddit, title, description, options, end_time)
    VALUES ($1, $2, $3, $4, $5, $6)
	`

	GetAllPollsQuery = `
		SELECT COUNT (*) OVER () AS total,
    	p.id, 
    	p.reddit_uid, 
    	p.subreddit, 
    	p.title, 
    	p.description, 
    	p.options, 
    	p.start_time, 
    	p.end_time, 
    	p.is_active, 
    	u.username, 
    	u.avatar,
    	(SELECT COUNT(*) FROM poll_votes pv WHERE pv.poll_id = p.id) AS total_votes,
    	(
        	SELECT jsonb_object_agg(o.id, COALESCE(v.vote_count, 0))
        	FROM jsonb_array_elements(p.options) WITH ORDINALITY AS o(option, id)
        	LEFT JOIN (
            	SELECT option_id, COUNT(*) AS vote_count
            	FROM poll_votes
            	WHERE poll_id = p.id
            	GROUP BY option_id
        	) v ON v.option_id = o.id
    	) AS vote_counts
	FROM polls p
	JOIN users u ON p.reddit_uid = u.reddit_uid
	WHERE p.subreddit = $1
	ORDER BY p.start_time DESC  
	LIMIT $2
	OFFSET $3
	`

	GetAllPollsQuerySigned = `
		SELECT 
    	COUNT(*) OVER () AS total,
    	p.id, 
    	p.reddit_uid, 
    	p.subreddit, 
    	p.title, 
    	p.description, 
    	p.options, 
    	p.start_time, 
    	p.end_time, 
    	p.is_active, 
    	u.username, 
    	u.avatar,
    	(SELECT COUNT(*) FROM poll_votes pv WHERE pv.poll_id = p.id) AS total_votes,
    	(
        	SELECT jsonb_object_agg(o.id, COALESCE(v.vote_count, 0))
        	FROM jsonb_array_elements(p.options) WITH ORDINALITY AS o(option, id)
        	LEFT JOIN (
            	SELECT option_id, COUNT(*) AS vote_count
            	FROM poll_votes
            	WHERE poll_id = p.id
            	GROUP BY option_id
        	) v ON v.option_id = o.id
    	) AS vote_counts,
    	(
        	SELECT pv.option_id
        	FROM poll_votes pv
        	WHERE pv.poll_id = p.id AND pv.reddit_uid = $1
    	) AS user_vote
	FROM polls p
	JOIN users u ON p.reddit_uid = u.reddit_uid
	WHERE p.subreddit = $2
	ORDER BY p.start_time DESC  
	LIMIT $3
	OFFSET $4
	`

	GetPollByIDQuery = `
	SELECT 
    	p.id,
    	p.reddit_uid,
    	p.subreddit,
    	p.title,
    	p.description,
    	p.options,
    	p.start_time,
    	p.end_time,
    	p.is_active, 
    	u.username,
    	u.avatar,
    	(SELECT COUNT(*) FROM poll_votes pv WHERE pv.poll_id = p.id) AS total_votes,
    	(
        	SELECT jsonb_object_agg(o.id, COALESCE(v.vote_count, 0))
        	FROM jsonb_array_elements(p.options) WITH ORDINALITY AS o(option, id)
        	LEFT JOIN (
            	SELECT option_id, COUNT(*) AS vote_count
            	FROM poll_votes
            	WHERE poll_id = p.id
            	GROUP BY option_id
        	) v ON v.option_id = o.id
    	) AS vote_counts
	FROM polls p
	JOIN users u ON p.reddit_uid = u.reddit_uid
	WHERE p.id = $1
	`

	PollLimitForUserQuery = `
	select exists (select 1 from polls where reddit_uid = $1 and start_time > now() - interval '6 hours')
	`

	CreatePollVoteQuery = `
	insert into poll_votes (poll_id, reddit_uid, option_id) values ($1, $2, $3)
	on conflict (poll_id, reddit_uid) 
	do update set 
    	option_id = EXCLUDED.option_id, 
    	created_at = NOW()
	where 
    	poll_votes.created_at > NOW() - INTERVAL '15 minutes'
	`

	DeletePollByCreatorQuery = `
	delete from polls where id = $1 and reddit_uid = $2
	`
)
