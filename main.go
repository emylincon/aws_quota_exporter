package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/emylincon/aws_quota_exporter/pkg"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/slog"
)

// values populated by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "2023-09-03T17:54:45Z"
)

func closeHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		slog.Warn("Shutting down", "signal", "Keyboard Interrupt", "input", "Ctrl+C")
		os.RemoveAll(pkg.CacheFolder)
		os.Exit(0)
	}()
}

func printVersion() {
	dt, err := time.Parse(time.RFC3339, date)
	if err == nil {
		date = dt.Format(time.UnixDate)
	}
	appversion := struct {
		App       string
		Version   string
		Date      string
		Platform  string
		Commit    string
		GoVersion string
	}{
		App:       "AWS Quota Exporter (AQE)",
		Version:   version,
		Date:      date,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Commit:    commit,
		GoVersion: runtime.Version(),
	}
	fmt.Println(awsutil.Prettify(appversion))
}

func main() {
	var (
		configFile    = flag.String("config.file", "/etc/aqe/config.yml", "Path to configuration file.")
		logFormatType = flag.String("log.format", "text", "Format of log messages (text or json).")
		logFolder     = flag.String("log.folder", "stdout", "Folder to store logfiles. logs to stdout if not specified.")
		logLevel      = flag.String("log.level", "INFO", "Log level to log from (DEBUG|INFO|WARN|ERROR).")
		promPort      = flag.Int("prom.port", 10100, "port to expose prometheus metrics.")
		cacheDuration = flag.Duration("cache.duration", 300, "cache expiry time (seconds).")
		Version       = flag.Bool("version", false, "Display aqe version")
	)
	flag.Parse()

	if *Version {
		printVersion()
		os.Exit(0)
	}
	// create logger
	logger := pkg.NewLogger(*logFormatType, *logFolder, *logLevel).With("version", version)
	slog.SetDefault(logger)
	start := time.Now()
	slog.Info("Initializing AWS Quota Exporter")

	// Handle keyboard interrupt
	closeHandler()

	// Make Prometheus client aware of our collectors.
	qcl, err := pkg.NewQuotaConfig(*configFile)
	if err != nil {
		slog.Error(fmt.Sprintf("Error parsing '%s'", *configFile), "error", err)
		return
	}
	s, err := pkg.NewScraper()
	if err != nil {
		slog.Error("Error creating scraper", "error", err)
		return
	}

	reg := prometheus.NewRegistry()
	slog.Info("Registering scrappers")
	for _, job := range qcl.Jobs {

		pc := pkg.NewPrometheusCollector(s.CreateScraper(job, cacheDuration))
		err = reg.Register(pc)
		if err != nil {
			slog.Error("Failed to register metrics: "+err.Error(), "serviceCode", job.ServiceCode, "regions", job.Regions, "role", job.Role)
		}
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

	slog.Info("Initialization of AWS Quota Exporter completed successfully", "duration", time.Since(start))

	// Start listening for HTTP connections.
	port := fmt.Sprintf(":%d", *promPort)
	slog.Info("Starting AWS Quota Exporter", "address", fmt.Sprintf("%v/metrics", port))
	if err := http.ListenAndServe(port, mux); err != nil {
		slog.Error("Cannot start AWS Quota Exporter", "error", err)
	}
}
