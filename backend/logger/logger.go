package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
)

func Init(logDir string) error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log dir: %w", err)
	}

	logFile := filepath.Join(logDir, fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02")))
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	multi := io.MultiWriter(os.Stdout, f)

	flags := log.Ldate | log.Ltime | log.Lmicroseconds
	Info = log.New(multi, "[INFO]  ", flags)
	Warn = log.New(multi, "[WARN]  ", flags)
	Error = log.New(multi, "[ERROR] ", flags)

	Info.Printf("Logger initialized, log file: %s", logFile)
	return nil
}
