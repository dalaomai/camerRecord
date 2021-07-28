package logging

import (
	"os"

	logging "github.com/op/go-logging"
)

// Logger ...
var logger *logging.Logger

var format = logging.MustStringFormatter(
	`%{time:2006-1-2 15:04:05.000} - %{pid} - %{longfile} - %{level:.4s} - %{message}`,
)

// InitLogger ...
func initLogger() {
	os.Mkdir(".log", os.ModePerm)
	logFile, err := os.OpenFile(".log/log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}

	backend := logging.NewLogBackend(logFile, "", 0)
	formatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(formatter)

	logger = logging.MustGetLogger("camerRecord")
}

// GetLogger ...
func GetLogger() *logging.Logger {
	if logger == nil {
		initLogger()
	}
	return logger
}
