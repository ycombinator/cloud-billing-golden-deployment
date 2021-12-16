package logging

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

func init() {
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	Logger = l
	defer Logger.Sync()
}
