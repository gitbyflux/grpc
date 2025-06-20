package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gitbyflux/grpcpractice/internal/domain/models"
	"github.com/gitbyflux/grpcpractice/internal/lib/logger/sl"
	"github.com/gitbyflux/grpcpractice/internal/storage"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func (s *Storage) App(ctx context.Context,
	appID int,
) (models.App, error) {
	const op = "storage.sqlite.App"

	stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = ?")
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

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, sl.Wrap(op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context,
	email string,
	passHash []byte,
) (int64, error) {
	const op = "storage.sqlite.SaveUser"

	stmt, err := s.db.Prepare("INSERT INTO users(email, pass_Hash) VALUES(?, ?)")
	if err != nil {
		return 0, sl.Wrap(op, err)
	}

	res, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {
		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) &&
			sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, sl.Wrap(op, storage.ErrUserExists)
		}

		return 0, sl.Wrap(op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, sl.Wrap(op, err)
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context,
	email string,
) (models.User, error) {
	const op = "storage.sqlite.User"

	stmt, err := s.db.Prepare("SELECT id, email, pass_hash FROM users WHERE email = ?")
	if err != nil {
		return models.User{}, sl.WrapMsg(op, "prepare statement", err)
	}

	row := stmt.QueryRowContext(ctx, email)

	var user models.User

	err = row.Scan(&user.ID, &user.Email, &user.PassHash)
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
	const op = "storage.sqlite.User"

	stmt, err := s.db.Prepare("SELECT is_admin FROM users WHERE id = ?")
	if err != nil {
		return false, sl.WrapMsg(op, "prepare statement", err)
	}

	row := stmt.QueryRowContext(ctx, userID)

	var isAdmin bool

	err = row.Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, sl.WrapMsg(op, "execute statement", storage.ErrUserNotFound)
		}
		return false, sl.WrapMsg(op, "execute statement", err)
	}

	return isAdmin, nil
}
