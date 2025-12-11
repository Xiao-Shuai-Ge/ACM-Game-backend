package grpc

import (
	"context"
	"net"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	conf "acmgame-backend/internal/common/config"
	"acmgame-backend/internal/common/server/grpc/interceptors"
)

// ServiceRegister allows services to register themselves with the gRPC server
type ServiceRegister func(*grpc.Server)

type ServerParams struct {
	fx.In
	Logger    *zap.Logger
	Config    *conf.Config
	Registers []ServiceRegister `group:"grpc_registers"`
}

func startGRPC(p ServerParams, lc fx.Lifecycle) {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptors.UnaryRecovery(p.Logger),
			interceptors.UnaryLogger(p.Logger),
		),
	}

	srv := grpc.NewServer(opts...)

	// Register all injected services
	for _, reg := range p.Registers {
		reg(srv)
	}

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
