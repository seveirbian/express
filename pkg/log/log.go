package log

import (
	"fmt"
	"go.uber.org/zap"
)

var (
	Logger      *zap.Logger
	SugarLogger *zap.SugaredLogger
)

func init() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("[logger] failed to new zap log for %v\n", err)
	}

	Logger = logger
	SugarLogger = logger.Sugar()
}
