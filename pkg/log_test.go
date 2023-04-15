package pkg

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/exp/slog"
)

func TestNewLogger(t *testing.T) {
	type args struct {
		formatType string
		logFolder  string
		logLevel   string
	}
	var LogLevel = new(slog.LevelVar)
	tests := []struct {
		name string
		args args
		want *slog.Logger
	}{
		{
			name: "json_handler_debug",
			args: args{
				formatType: "json",
				logFolder:  "stdout",
				logLevel:   "debug",
			},
			want: slog.New(slog.HandlerOptions{Level: LogLevel}.NewJSONHandler(os.Stdout)),
		},
		{
			name: "json_handler_info",
			args: args{
				formatType: "json",
				logFolder:  "stdout",
				logLevel:   "info",
			},
			want: slog.New(slog.HandlerOptions{Level: LogLevel}.NewJSONHandler(os.Stdout)),
		},
		{
			name: "text_handler_debug",
			args: args{
				formatType: "text",
				logFolder:  "stdout",
				logLevel:   "debug",
			},
			want: slog.New(slog.HandlerOptions{Level: LogLevel}.NewTextHandler(os.Stdout)),
		},
		{
			name: "text_handler_info",
			args: args{
				formatType: "text",
				logFolder:  "stdout",
				logLevel:   "info",
			},
			want: slog.New(slog.HandlerOptions{Level: LogLevel}.NewTextHandler(os.Stdout)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch strings.ToUpper(tt.args.logLevel) {
			case "DEBUG":
				LogLevel.Set(slog.LevelDebug)
			case "INFO":
				LogLevel.Set(slog.LevelInfo)
			case "WARN":
				LogLevel.Set(slog.LevelWarn)
			case "ERROR":
				LogLevel.Set(slog.LevelError)
			}
			if got := NewLogger(tt.args.formatType, tt.args.logFolder, tt.args.logLevel); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewLogWriter(t *testing.T) {
	type args struct {
		folder string
	}
	tests := []struct {
		name    string
		args    args
		want    *LogWriter
		wantErr bool
	}{
		{
			name: "test empty folder",
			args: args{
				folder: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLogWriter(tt.args.folder)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLogWriter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLogWriter() = %v, want %v", got, tt.want)
			}
		})
	}
}
