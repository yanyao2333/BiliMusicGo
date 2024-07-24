package main

import (
	"github.com/CuteReimu/bilibili/v2"
	"go.uber.org/zap"
)

func Logger() *zap.SugaredLogger {
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()
	return zapLogger.Sugar()
}

func main() {
	logger := Logger()
	bili := BiliMonitorStruct{BiliClient: bilibili.New(), logger: logger}
	bili.InitListener()
}
