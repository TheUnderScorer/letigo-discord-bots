package logging

import (
	"app/env"
	"app/metadata"
	"fmt"
	"go.uber.org/zap"
	"time"
)

var logger *zap.Logger

func Init() {
	var err error
	var cfg zap.Config

	logFileId := fmt.Sprintf("%s-%s", metadata.GetVersion(), time.Now().Format(time.RFC3339))
	logFile := fmt.Sprintf("./logs/%s.log", logFileId)

	if env.IsProd() {
		cfg = zap.NewProductionConfig()
		cfg.OutputPaths = []string{logFile, "stderr"}
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
