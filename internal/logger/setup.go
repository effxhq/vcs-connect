package logger

import (
	"go.uber.org/zap"
	"log"
)

func MustSetup() *zap.Logger {
	logger, err := Setup()
	if err != nil {
		log.Fatal(err.Error())
	}
	return logger
}

func Setup() (*zap.Logger, error) {
	return zap.Config{
		Development:      false,
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
	}.Build()
}
