package grpc

import (
    "context"

    platformv1 "acmgame-backend/api/gen/go/platform/v1"
    "acmgame-backend/internal/platform/service"
    "go.uber.org/fx"
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

var Module = fx.Options(
    fx.Provide(NewPlatformHandler),
)

