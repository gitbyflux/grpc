package psql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gitbyflux/grpcpractice/internal/domain/models"
	"github.com/gitbyflux/grpcpractice/internal/lib/logger/sl"
	"github.com/gitbyflux/grpcpractice/internal/storage"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func (s *Storage) App(ctx context.Context,
	appID int,
) (models.App, error) {
	const op = "storage.psql.App"

	stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = $1")
	if err != nil {
		return models.App{}, sl.WrapMsg(op, "prepare statement", err)
	}

	row := stmt.QueryRowContext(ctx, appID)

	var app models.App

	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, sl.WrapMsg(op, "execute statement", storage.ErrAppNotFound)
		}
		return models.App{}, sl.WrapMsg(op, "execute statement", err)
	}

	return app, nil
}

func New(dsn string) (*Storage, error) {
	const op = "storage.psql.New"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, sl.Wrap(op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context,
	email string,
	passHash []byte,
) (int64, error) {
	const op = "storage.psql.SaveUser"

	var id int64
	err := s.db.QueryRowContext(ctx,
		"INSERT INTO users(email, pass_hash) VALUES($1, $2) RETURNING id",
		email, passHash,
	).Scan(&id)
	if err != nil {
		var pgErr *pq.Error

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" {
			return 0, sl.Wrap(op, storage.ErrUserExists)
		}
		return 0, sl.Wrap(op, err)
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context,
	email string,
) (models.User, error) {
	const op = "storage.psql.User"

	var user models.User
	err := s.db.QueryRowContext(ctx,
		"SELECT id, email, pass_hash FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, sl.WrapMsg(op, "execute statement", storage.ErrUserNotFound)
		}
		return models.User{}, sl.WrapMsg(op, "execute statement", err)
	}

	return user, nil
}

func (s *Storage) IsAdmin(ctx context.Context,
	userID int64,
) (bool, error) {
	const op = "storage.psql.User"

	var isAdmin bool
	err := s.db.QueryRowContext(ctx,
		"SELECT is_admin FROM users WHERE id = $1",
		userID).Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, sl.WrapMsg(op, "execute statement", storage.ErrUserNotFound)
		}
		return false, sl.WrapMsg(op, "execute statement", err)
	}

	return isAdmin, nil
}
