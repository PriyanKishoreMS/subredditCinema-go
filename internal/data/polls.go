package data

import (
	"encoding/json"
	"time"

	pgx "github.com/jackc/pgx/v5/pgxpool"
)

type PollsModel struct {
	DB *pgx.Pool
}

type PollOption struct {
	ID     int    `json:"id" validate:"required"`
	Text   string `json:"text" validate:"required"`
	ImgURL string `json:"img_url,omitempty"`
}

type Poll struct {
	RedditUID    string          `json:"reddit_uid" validate:"required"`
	Subreddit    string          `json:"subreddit" validate:"required"`
	Title        string          `json:"title" validate:"required"`
	Description  string          `json:"description"`
	Options      json.RawMessage `json:"options" validate:"required"`
	VotingMethod string          `json:"voting_method"`
	EndTime      time.Time       `json:"end_time" validate:"required"`
}

type PollDataResponse struct {
	ID           int             `json:"id"`
	RedditUID    string          `json:"reddit_uid"`
	Subreddit    string          `json:"subreddit"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	Options      json.RawMessage `json:"options"`
	VotingMethod string          `json:"voting_method"`
	StartTime    time.Time       `json:"start_time"`
	EndTime      time.Time       `json:"end_time"`
	IsActive     bool            `json:"is_active"`
	UserName     string          `json:"user_name"`
	UserAvatar   string          `json:"user_avatar"`
	TotalVotes   int             `json:"total_votes"`
	VoteCount    json.RawMessage `json:"vote_count"`
}

func (p PollsModel) InsertNewPoll(poll *Poll, options []PollOption) error {
	ctx, cancel := Handlectx()
	defer cancel()

	query := CreatePollsQuery

	_, err := p.DB.Exec(ctx, query, poll.RedditUID, poll.Subreddit, poll.Title, poll.Description, options, poll.VotingMethod, poll.EndTime)
	if err != nil {
		return err
	}

	return nil
}

func (p PollsModel) PollLimitForUser(redditUID string) (bool, error) {
	ctx, cancel := Handlectx()
	defer cancel()

	query := PollLimitForUserQuery

	var exists bool
	err := p.DB.QueryRow(ctx, query, redditUID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (p PollsModel) GetAllPolls(sub string, filters Filters) ([]PollDataResponse, Metadata, error) {
	ctx, cancel := Handlectx()
	defer cancel()

	query := GetAllPollsQuery

	rows, err := p.DB.Query(ctx, query, sub, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var polls []PollDataResponse
	totalRecords := 0
	for rows.Next() {
		var poll PollDataResponse
		err := rows.Scan(&totalRecords, &poll.ID, &poll.RedditUID, &poll.Subreddit, &poll.Title, &poll.Description, &poll.Options, &poll.VotingMethod, &poll.StartTime, &poll.EndTime, &poll.IsActive, &poll.UserName, &poll.UserAvatar, &poll.TotalVotes, &poll.VoteCount)
		if err != nil {
			return nil, Metadata{}, err
		}
		polls = append(polls, poll)
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return polls, metadata, nil
}

func (p PollsModel) GetPollByID(pollID int) (*PollDataResponse, error) {
	ctx, cancel := Handlectx()
	defer cancel()

	query := GetPollByIDQuery

	var poll PollDataResponse
	err := p.DB.QueryRow(ctx, query, pollID).Scan(&poll.ID, &poll.RedditUID, &poll.Subreddit, &poll.Title, &poll.Description, &poll.Options, &poll.VotingMethod, &poll.StartTime, &poll.EndTime, &poll.IsActive, &poll.UserName, &poll.UserAvatar, &poll.TotalVotes, &poll.VoteCount)
	if err != nil {
		return nil, err
	}

	return &poll, nil
}

func (p PollsModel) CreatePollVote(pollID int, redditUID string, optionID int) error {
	ctx, cancel := Handlectx()
	defer cancel()

	query := CreatePollVoteQuery

	_, err := p.DB.Exec(ctx, query, pollID, redditUID, optionID)
	if err != nil {
		return err
	}

	return nil
}

func (p PollsModel) DeletePollByCreator(pollID int, redditUID string) error {
	ctx, cancel := Handlectx()
	defer cancel()

	query := DeletePollByCreatorQuery

	_, err := p.DB.Exec(ctx, query, pollID, redditUID)
	if err != nil {
		return err
	}

	return nil
}
