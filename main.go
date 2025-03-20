// Package main implements the AWS Quota Exporter (AQE), a tool for exporting AWS service quotas
// as Prometheus metrics. The application provides functionality to scrape AWS service quota
// information, expose it as Prometheus metrics, and serve it over HTTP.
//
// The main package includes the following features:
// - Command-line flags for configuration, including log settings, Prometheus port, and cache duration.
// - Graceful shutdown handling using OS signals.
// - Build information exposure as Prometheus metrics and version display.
// - Integration with Prometheus for metrics collection and HTTP serving.
//
// The application is designed to be extensible and configurable, allowing users to customize
// logging, caching, and metric collection behavior.
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

type buildInfo struct {
	App       string
	Version   string
	Date      string
	Platform  string
	Commit    string
	GoVersion string
}

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

// getbuildInfo constructs and returns a buildInfo struct containing metadata
// about the application build, such as the application name, version, build date,
// platform, commit hash, and Go runtime version. The build date is parsed and
// reformatted to a human-readable format if it adheres to the RFC3339 standard.
// If the date parsing fails, the original date string is used.
func getbuildInfo() buildInfo {
	dt, err := time.Parse(time.RFC3339, date)
	if err == nil {
		date = dt.Format(time.UnixDate)
	}
	return buildInfo{
		App:       "AWS Quota Exporter (AQE)",
		Version:   version,
		Date:      date,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Commit:    commit,
		GoVersion: runtime.Version(),
	}
}

func printVersion() {
	appversion := getbuildInfo()
	fmt.Println(awsutil.Prettify(appversion))
}

// buildInfoMetrics generates a slice of Prometheus metrics containing build information
// about the application. It retrieves build details such as the application name, version,
// build date, platform, commit hash, and Go version, and packages them into a Prometheus
// metric with the name "aqe_build_info".
//
// Returns:
//   - A slice of pointers to PrometheusMetric containing the build information.
//   - An error if any issue occurs during the metric creation process.
func buildInfoMetrics() ([]*pkg.PrometheusMetric, error) {
	appversion := getbuildInfo()
	var metrics []*pkg.PrometheusMetric

	labels := map[string]string{
		"app":        appversion.App,
		"version":    appversion.Version,
		"build_date": appversion.Date,
		"platform":   appversion.Platform,
		"commit":     appversion.Commit,
		"go_version": appversion.GoVersion,
	}

	metrics = append(metrics, &pkg.PrometheusMetric{
		Name:   "aqe_build_info",
		Labels: labels,
		Value:  1,
		Desc:   "AQE Build information",
	})

	return metrics, nil
}

func main() {
	var (
		configFile      = flag.String("config.file", "/etc/aqe/config.yml", "Path to configuration file.")
		logFormatType   = flag.String("log.format", "text", "Format of log messages (text or json).")
		logFolder       = flag.String("log.folder", "stdout", "Folder to store logfiles. logs to stdout if not specified.")
		logLevel        = flag.String("log.level", "INFO", "Log level to log from (DEBUG|INFO|WARN|ERROR).")
		promPort        = flag.Int("prom.port", 10100, "Port to expose prometheus metrics.")
		cacheDuration   = flag.Duration("cache.duration", 300*time.Second, "Cache expiry time.")
		cacheServeStale = flag.Bool("cache.serve-stale", false, "Serve stale cache data if available during cache refresh.")
		collectUsage    = flag.Bool("collect.usage", false, "Collect quotas usage where available (NOTE: CloudWatch calls aren't free)")
		Version         = flag.Bool("version", false, "Display aqe version")
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

		pc := pkg.NewPrometheusCollector(s.CreateScraper(job, cacheDuration, *cacheServeStale, *collectUsage))
		err = reg.Register(pc)
		if err != nil {
			slog.Error("Failed to register metrics: "+err.Error(), "serviceCode", job.ServiceCode, "regions", job.Regions, "role", job.Role)
		}
	}

	reg.Register(pkg.NewPrometheusCollector(buildInfoMetrics))

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
