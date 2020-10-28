package logger

import (
	"log"

	"go.uber.org/zap"
)

// MustSetup forces a setup of the logger and fails the program if an error occurs.
func MustSetup() *zap.Logger {
	logger, err := Setup()
	if err != nil {
		log.Fatal(err.Error())
	}
	return logger
}

// Setup attempts to initialize a logger for this program.
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
