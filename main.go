package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/emylincon/aws_quota_exporter/pkg"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/slog"
)

func closeHandler(logger *slog.Logger) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		logger.Warn("Shutting down", "signal", "Keyboard Interrupt", "input", "Ctrl+C")
		os.Exit(0)
	}()
}

// NewLogger returns a logger
func NewLogger(formatType string) *slog.Logger {
	if formatType == "json" {
		return slog.New(slog.NewJSONHandler(os.Stdout))
	}
	return slog.New(slog.NewTextHandler(os.Stdout))
}

func main() {
	var (
		configFile    = flag.String("config.file", "config.yaml", "Path to configuration file. Defaults to config.yaml")
		logFormatType = flag.String("log.format", "text", "Format of log messages (text or json). Defaults to text")
		promPort      = flag.Int("prom.port", 10100, "port to expose prometheus metrics, Defaults to 10100")
	)
	flag.Parse()

	version := "0.0.0"

	logger := NewLogger(*logFormatType).With("version", version)

	logger.Info("Initializing AWS Quota Exporter")

	// Handle keyboard interrupt
	closeHandler(logger)

	// Make Prometheus client aware of our collectors.
	qcl, err := pkg.NewQuotaConfig(*configFile)
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing '%s'", *configFile), "error", err)
		return
	}
	s, err := pkg.NewScraper()
	if err != nil {
		logger.Error("Error creating scraper", "error", err)
		return
	}

	reg := prometheus.NewRegistry()
	for _, qc := range qcl.Jobs {
		pc := pkg.NewPrometheusCollector(logger, s.CreateScraper(qc.Regions, qc.ServiceCode))
		reg.MustRegister(pc)
	}

	mux := http.NewServeMux()
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	mux.Handle("/metrics", promHandler)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(fmt.Sprintf(`<html>
    <head><title>AWS Quota Exporter</title></head>
    <body>
    <h1>AWS Quota Exporter</h1>
		Version: %s
    <p><a href="/metrics">Metrics</a></p>
    </body>
    </html>`, version)))
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Start listening for HTTP connections.
	port := fmt.Sprintf(":%d", *promPort)
	logger.Info("Starting AWS Quota Exporter", "address", fmt.Sprintf(":%v/metrics", port))
	if err := http.ListenAndServe(port, mux); err != nil {
		logger.Error("Cannot start AWS Quota Exporter", "error", err)
	}
}
