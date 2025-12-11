package main

import (
	"context"

	"go.uber.org/fx"

	"acmgame-backend/internal/common/config"
	"acmgame-backend/internal/common/log"
	grpcserver "acmgame-backend/internal/common/server/grpc"
	httpserver "acmgame-backend/internal/common/server/http"
	platformgrpc "acmgame-backend/internal/platform/handler/grpc"
	platformsrv "acmgame-backend/internal/platform/service"
)

func main() {
	app := fx.New(
		config.Module,
		log.Module,
		platformsrv.Module,
		platformgrpc.Module,
		grpcserver.Module,
		httpserver.Module,
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error { return nil },
				OnStop:  func(ctx context.Context) error { return nil },
			})
		}),
	)
	app.Run()
}
