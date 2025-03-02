package logger

import "go.uber.org/zap"

var ZapLogger *zap.Logger

func InitLogger() {
	cfg := zap.NewProductionConfig()
	ZapLogger, _ = cfg.Build()
}

func Error(){
	
}