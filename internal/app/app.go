package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/gitbyflux/grpcpractice/internal/app/grpc"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {
	grpcApp := grpcapp.New(log, grpcPort)
	return &App{
		GRPCSrv: grpcApp,
	}
}
