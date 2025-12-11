package grpc

import (
	"context"

	platformv1 "acmgame-backend/api/gen/go/platform/v1"
	grpcserver "acmgame-backend/internal/common/server/grpc"
	"acmgame-backend/internal/platform/service"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type PlatformHandler struct {
	svc *service.Service
	platformv1.UnimplementedPlatformServiceServer
}

func NewPlatformHandler(svc *service.Service) *PlatformHandler {
	return &PlatformHandler{svc: svc}
}

func (h *PlatformHandler) Ping(ctx context.Context, req *platformv1.PingRequest) (*platformv1.PingResponse, error) {
	return &platformv1.PingResponse{Message: h.svc.Ping()}, nil
}

// RegisterService wraps the gRPC registration
func RegisterService(h *PlatformHandler) grpcserver.ServiceRegister {
	return func(srv *grpc.Server) {
		platformv1.RegisterPlatformServiceServer(srv, h)
	}
}

var Module = fx.Options(
	fx.Provide(NewPlatformHandler),
	fx.Provide(
		fx.Annotate(
			RegisterService,
			fx.ResultTags(`group:"grpc_registers"`),
		),
	),
)
