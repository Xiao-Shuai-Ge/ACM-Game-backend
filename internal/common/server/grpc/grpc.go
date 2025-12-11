package grpc

import (
	"net"

	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	platformv1 "acmgame-backend/api/gen/go/platform/v1"
	conf "acmgame-backend/internal/common/config"
	platformgrpc "acmgame-backend/internal/platform/handler/grpc"
)

type ServerParams struct {
	fx.In
	Logger  *zap.Logger
	Config  *conf.Config
	Handler *platformgrpc.PlatformHandler
}

func startGRPC(p ServerParams, lc fx.Lifecycle) {
	srv := grpc.NewServer()
	platformv1.RegisterPlatformServiceServer(srv, p.Handler)
	reflection.Register(srv)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", p.Config.Server.GRPC.Addr)
			if err != nil {
				p.Logger.Error("grpc listen error", zap.Error(err))
				return err
			}
			go func() {
				p.Logger.Info("grpc server started", zap.String("addr", p.Config.Server.GRPC.Addr))
				if err := srv.Serve(ln); err != nil {
					p.Logger.Error("grpc serve error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			srv.GracefulStop()
			return nil
		},
	})
}

var Module = fx.Options(
	fx.Invoke(startGRPC),
)
