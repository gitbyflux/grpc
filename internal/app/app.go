package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/gitbyflux/grpcpractice/internal/app/grpc"
	"github.com/gitbyflux/grpcpractice/internal/services/auth"
	"github.com/gitbyflux/grpcpractice/internal/storage/psql"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	dsn string,
	tokenTTL time.Duration,
) *App {
	//storage, err := sqlite.New(storagePath)
	storage, err := psql.New(dsn)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
