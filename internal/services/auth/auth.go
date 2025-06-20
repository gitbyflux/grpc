package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gitbyflux/grpcpractice/internal/domain/models"
	"github.com/gitbyflux/grpcpractice/internal/lib/jwt"
	"github.com/gitbyflux/grpcpractice/internal/lib/logger/sl"
	"github.com/gitbyflux/grpcpractice/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
)

func New(log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		usrSaver:    userSaver,
		usrProvider: userProvider,
		log:         log,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found")

			return "", sl.Wrap(op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return "", sl.Wrap(op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", sl.Wrap(op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", sl.Wrap(op, err)
	}

	log.Info("user logger in successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return "", sl.Wrap(op, err)
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context,
	email string,
	pass string,
) (int64, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))

		return 0, sl.Wrap(op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exist", sl.Err(err))

			return 0, sl.Wrap(op, ErrInvalidCredentials)
		}
		log.Error("failed to save user", sl.Err(err))

		return 0, sl.Wrap(op, err)
	}

	log.Info("user registered")

	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context,
	userID int64,
) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", int64(userID)),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("user not found", sl.Err(err))

			return false, sl.Wrap(op, ErrInvalidAppID)
		}
		return false, sl.Wrap(op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
