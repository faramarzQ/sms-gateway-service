package logger

import (
	"fmt"
	"go.uber.org/zap"
	"sync"
)

var (
	Logger *zap.Logger
	once   sync.Once
)

func Init(appName string) error {
	var err error

	once.Do(func() {
		cfg := zap.NewProductionConfig()
		logFile := fmt.Sprintf("logs/%s.log", appName)
		cfg.OutputPaths = []string{logFile}

		Logger, err = cfg.Build()
	})

	return err
}
