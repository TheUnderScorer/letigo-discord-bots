package logging

import (
	"app/env"
	"go.uber.org/zap"
)

var logger *zap.Logger

func Init() {
	var err error
	var cfg zap.Config

	if env.IsProd() {
		cfg = zap.NewProductionConfig()
		cfg.OutputPaths = []string{"out.log", "stderr"}
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	logger, err = cfg.Build()
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
