package logger

import (
	"go.uber.org/zap"
)

var log *zap.Logger

func init() {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	log = l
}

func Error(err error) {
	defer log.Sync()
	sugar := log.Sugar()
	sugar.Error(err)
}

func Infof(template string, args ...interface{}) {
	defer log.Sync()
	sugar := log.Sugar()
	sugar.Infof(template, args)
}

func Debug(msg string, fields ...zap.Field) {
	defer log.Sync()
	sugar := log.Sugar()
	sugar.Debug(msg, fields)
}

func Debugf(template string, args ...interface{}) {
	defer log.Sync()
	sugar := log.Sugar()
	sugar.Debugf(template, args)
}
