package logger

import "go.uber.org/zap"

// type Logger struct {
// 	*zap.SugaredLogger
// }

// var Log *zap.Logger
var Log *zap.SugaredLogger

func InitLogger() {
	logger, _ := zap.NewDevelopment()
	// Log = logger
	Log = logger.Sugar()
}
