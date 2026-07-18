package logger

import (
	"go.uber.org/zap"
	"sync"
)

var (
	Logger *zap.Logger
	once   sync.Once
)

func Init() error {
	var err error

	once.Do(func() {
		cfg := zap.NewProductionConfig()
		cfg.OutputPaths = []string{"logs/app.log"}
		cfg.ErrorOutputPaths = []string{"logs/error.log"}

		Logger, err = cfg.Build()
	})

	return err
}
