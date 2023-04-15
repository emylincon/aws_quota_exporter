package pkg

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/exp/slog"
)

// NewLogWriter creates a new LogWriter
func NewLogWriter(folder string) (*LogWriter, error) {
	// check if cache folder exists
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		err = os.MkdirAll(folder, 0755)
		if err != nil {
			return nil, err
		}
	}
	lw := LogWriter{
		logFolder: folder,
	}
	lw.setlogFile()
	return &lw, nil

}

// LogWriter is a wrapper around os.File
type LogWriter struct {
	logFolder string
	logFile   string
	expires   time.Time
}

// Write implements io.Writer
func (lw *LogWriter) Write(b []byte) (n int, err error) {
	if time.Now().After(lw.expires) {
		lw.setlogFile()
	}
	f, err := os.OpenFile(lw.getlogFile(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		fmt.Printf("Error opening log file - %s: error=%s\n", lw.getlogFile(), err)
		fmt.Println("Switching to default console logging")
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout)))
		slog.Info("Process recovered successfully from error while writing logs", "error", err)
	}
	defer f.Close()

	return f.Write(b)
}

func (lw *LogWriter) setlogFile() {
	lw.logFile = fmt.Sprintf("aws-quota-exporter-%s.log", time.Now().Format("02-01-2006"))
	lw.expires = time.Now().Add(time.Hour * 24)
}

func (lw *LogWriter) getlogFile() string {
	return fmt.Sprintf("%s/%s", lw.logFolder, lw.logFile)
}

// NewLogger returns a logger
func NewLogger(formatType, logFolder, logLevel string) *slog.Logger {
	LogLevel := new(slog.LevelVar)
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		LogLevel.Set(slog.LevelDebug)
	case "INFO":
		LogLevel.Set(slog.LevelInfo)
	case "WARN":
		LogLevel.Set(slog.LevelWarn)
	case "ERROR":
		LogLevel.Set(slog.LevelError)
	}
	logOptions := slog.HandlerOptions{Level: LogLevel}
	var logwriter io.Writer
	logwriter = os.Stdout
	if logFolder != "stdout" {
		writer, err := NewLogWriter(logFolder)
		if err != nil {
			slog.Error("Error creating log folder", "error", err)
			slog.Warn("Switching to writing logs to console")
		} else {
			logwriter = writer
		}

	}
	if formatType == "json" {
		handler := logOptions.NewJSONHandler(logwriter)
		return slog.New(handler)
	}
	handler := logOptions.NewTextHandler(logwriter)
	return slog.New(handler)
}
