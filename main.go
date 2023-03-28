package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/emylincon/aws_quota_exporter/pkg"
	"github.com/emylincon/aws_quota_exporter/pkg/scrape"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func closeHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C : Closed Gracefully")
		os.Exit(0)
	}()
}

func main() {
	var (
		configFile = flag.String("config.file", "config.yaml", "Path to configuration file")
		promPort   = flag.Int("prom.port", 10100, "port to expose prometheus metrics")
	)
	flag.Parse()
	// Handle keyboard interrupt
	closeHandler()

	version := "0.0.0"
	// Make Prometheus client aware of our collectors.
	fmt.Println("Config file:", *configFile)
	qcl, err := pkg.NewQuotaConfig(*configFile)
	if err != nil {
		fmt.Printf("Error parsing '%s': %s", *configFile, err)
		return
	}
	profile := "emeka"
	s, err := scrape.NewScraper(profile)
	if err != nil {
		fmt.Println("Error creating scrape:", err)
		return
	}

	reg := prometheus.NewRegistry()
	for _, qc := range qcl.Jobs {
		pc := pkg.NewPrometheusCollector(s.CreateScraper(qc.Regions, qc.ServiceCode))
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
	log.Printf("starting nginx exporter on %q/metrics", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("cannot start nginx exporter: %s", err)
	}
}
