package data

import (
	"fmt"
	"time"

	pgx "github.com/jackc/pgx/v5/pgxpool"
)

type TierlistsModel struct {
	DB *pgx.Pool
}

type TierData struct {
	Label string `json:"label" validate:"required"`
	Color string `json:"color" validate:"required"`
}

type TierListData struct {
	ID        int        `json:"id"`
	RedditUID string     `json:"reddit_uid"`
	Title     string     `json:"title" validate:"required"`
	Subreddit string     `json:"subreddit" validate:"required"`
	Tiers     []TierData `json:"tiers,omitempty" validate:"required,dive"`
	Urls      []string   `json:"urls,omitempty" validate:"required"`
	CreatedAt time.Time  `json:"created_at"`
	ImageURL  string     `json:"image_url,omitempty"`
	Avatar    string     `json:"avatar,omitempty"`
	Username  string     `json:"username,omitempty"`
}

func (t TierlistsModel) CreateNewTierListTemplate(reddit_uid string, tierListData TierListData) (err error) {
	ctx, cancel := Handlectx()
	defer cancel()

	tx, err := t.DB.Begin(ctx)
	if err != nil {
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

	query := CreateTierListQuery
	var tierListID int

	err = tx.QueryRow(ctx, query, reddit_uid, tierListData.Title, tierListData.Subreddit, tierListData.Tiers).Scan(&tierListID)
	if err != nil {
		return
	}

	for _, url := range tierListData.Urls {
		query = InsertImageQuery
		var imgID int
		err = tx.QueryRow(ctx, query, url).Scan(&imgID)
		if err != nil {
			return
		}

		_, err = tx.Exec(ctx, MapTierListImageQuery, tierListID, imgID)
		if err != nil {
			return
		}

	}

	return nil
}

func (t TierlistsModel) GetAllTierlists(sub string, filters Filters) ([]TierListData, Metadata, error) {
	ctx, cancel := Handlectx()
	defer cancel()

	var tierLists []TierListData
	totalRecords := 0

	query := GetAllTierListsQuery

	rows, err := t.DB.Query(ctx, query, sub, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	for rows.Next() {
		tierlist := new(TierListData)
		if err := rows.Scan(&tierlist.ID, &tierlist.RedditUID, &tierlist.Title, &tierlist.Subreddit, &tierlist.CreatedAt, &tierlist.ImageURL, &tierlist.Avatar, &tierlist.Username, &totalRecords); err != nil {
			return nil, Metadata{}, err
		}
		tierLists = append(tierLists, *tierlist)
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return tierLists, metadata, nil
}

func (t TierlistsModel) GetTierListByID(tierlistID int) (*TierListData, error) {
	ctx, cancel := Handlectx()
	defer cancel()

	query := GetTierListByIDQuery

	var tierList TierListData
	err := t.DB.QueryRow(ctx, query, tierlistID).Scan(&tierList.ID, &tierList.RedditUID, &tierList.Title, &tierList.Subreddit, &tierList.CreatedAt, &tierList.Tiers, &tierList.Urls)
	if err != nil {
		return nil, err
	}

	return &tierList, nil
}
