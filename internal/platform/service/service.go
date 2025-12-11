package service

import "go.uber.org/fx"

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) Ping() string { return "pong" }

var Module = fx.Options(
    fx.Provide(NewService),
)

