package logger

import (
	"errors"

	"go.uber.org/zap"
)

var (
	ErrIncorrectEnv = errors.New("incorrect env")
)

func New(env string) (*zap.Logger, error) {
	switch env {
	case "dev":
		return newDevLogger()
	case "prod":
		return newProdLogger()
	default:
		return nil, ErrIncorrectEnv
	}
}

// Пока не понимаю, что там указывать. Пусть будут 2 стандартных логгера
func newDevLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}

func newProdLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}
