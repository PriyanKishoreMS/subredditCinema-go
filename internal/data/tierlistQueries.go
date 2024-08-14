package data

const (
	CreateTierListQuery = `INSERT INTO tierlists (reddit_uid, title, subreddit, tiers) VALUES ($1, $2, $3, $4) RETURNING id`

	InsertImageQuery = `INSERT INTO  tierlist_images (url) 
	VALUES ($1) 
	ON CONFLICT(url) 
	DO UPDATE SET url = EXCLUDED.url
	RETURNING id	
	`

	MapTierListImageQuery = `INSERT INTO tierlist_images_map (tierlist_id, image_id) VALUES($1, $2)`

	GetAllTierListsQuery = `SELECT 
    DISTINCT ON (tl.id) 
    tl.id, tl.reddit_uid, tl.title, tl.subreddit, tl.created_at, ti.url,
    (SELECT COUNT(DISTINCT id) FROM tierlists WHERE subreddit=$1) AS total
	FROM tierlists tl
	INNER JOIN tierlist_images_map tim ON tl.id = tim.tierlist_id 
	INNER JOIN tierlist_images ti ON ti.id = tim.image_id
	WHERE tl.subreddit=$1
	ORDER BY tl.id, tl.created_at DESC, ti.url
	LIMIT $2 OFFSET $3`

	GetTierListByIDQuery = `SELECT tl.id, tl.reddit_uid, tl.title, tl.subreddit, tl.created_at, tl.tiers, 
       array_agg(ti.url) AS images
	FROM tierlists tl 
	INNER JOIN tierlist_images_map tim ON tl.id = tim.tierlist_id 
	INNER JOIN tierlist_images ti ON ti.id = tim.image_id 
	WHERE tl.id = $1
	GROUP BY tl.id, tl.reddit_uid, tl.title, tl.subreddit;`
)
