package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/gitbyflux/grpcpractice/internal/app/grpc"
	"github.com/gitbyflux/grpcpractice/internal/services/auth"
	"github.com/gitbyflux/grpcpractice/internal/storage/sqlite"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
