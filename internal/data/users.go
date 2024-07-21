package data

import pgx "github.com/jackc/pgx/v5/pgxpool"

type UserModel struct {
	DB *pgx.Pool
}

type User struct {
	ID        int
	RedditUID string
	Username  string
	Avatar    string
}

func (u UserModel) CheckUserExists(reddit_id string) (bool, error) {
	ctx, cancel := Handlectx()
	defer cancel()

	query := CheckUserExistsQuery

	var exists bool

	err := u.DB.QueryRow(ctx, query, reddit_id).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (u UserModel) InsertUser(username, avatar, reddit_id string) (User, error) {
	ctx, cancel := Handlectx()
	defer cancel()

	query := InsertUserQuery

	var user User

	row := u.DB.QueryRow(ctx, query, reddit_id, username, avatar)

	err := row.Scan(
		&user.ID,
		&user.RedditUID,
		&user.Username,
		&user.Avatar,
	)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
