package logger

import "go.uber.org/zap"

func New(level string) *zap.Logger {
	cfg := zap.NewProductionConfig()
	if level == "debug" {
		cfg = zap.NewDevelopmentConfig()
	}
	lg, _ := cfg.Build()
	return lg
}
