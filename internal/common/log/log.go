package log

import (
    "go.uber.org/fx"
    "go.uber.org/zap"
)

func NewLogger() (*zap.Logger, error) {
    cfg := zap.NewProductionConfig()
    return cfg.Build()
}

var Module = fx.Options(
    fx.Provide(NewLogger),
)

