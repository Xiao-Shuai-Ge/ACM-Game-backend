package service

import (
	"acmgame-backend/internal/platform/repository"

	"go.uber.org/fx"
)

type Service struct {
	userRepo repository.UserRepository
}

func NewService(userRepo repository.UserRepository) *Service {
	return &Service{
		userRepo: userRepo,
	}
}

func (s *Service) Ping() string { return "pong" }

var Module = fx.Options(
	fx.Provide(repository.NewUserRepository),
	fx.Provide(NewService),
)
