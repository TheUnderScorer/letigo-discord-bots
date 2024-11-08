package logging

import (
	"app/env"
	"go.uber.org/zap"
)

var logger *zap.Logger

func Init() {
	var err error

	if env.IsProd() {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		panic(err)
	}
}

func Get() *zap.Logger {
	if logger == nil {
		Init()
	}

	return logger
}
