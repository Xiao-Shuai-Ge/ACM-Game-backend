package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	platformv1 "acmgame-backend/api/gen/go/platform/v1"
	conf "acmgame-backend/internal/common/config"
)

type ServerParams struct {
	fx.In
	Logger *zap.Logger
	Config *conf.Config
}

func startHTTP(p ServerParams, lc fx.Lifecycle) {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	gwmux := runtime.NewServeMux()
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			opts := []grpc.DialOption{grpc.WithInsecure()}
			if err := platformv1.RegisterPlatformServiceHandlerFromEndpoint(ctx, gwmux, p.Config.Server.GRPC.Addr, opts); err != nil {
				p.Logger.Error("gateway register error", zap.Error(err))
				return err
			}
			r.Any("/api/v1/*any", gin.WrapH(http.StripPrefix("/api/v1", gwmux)))
			go func() {
				p.Logger.Info("http server started", zap.String("addr", p.Config.Server.HTTP.Addr))
				if err := r.Run(p.Config.Server.HTTP.Addr); err != nil && err != http.ErrServerClosed {
					p.Logger.Error("http serve error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error { return nil },
	})
}

var Module = fx.Options(
	fx.Invoke(startHTTP),
)
