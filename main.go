package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/emylincon/aws_quota_exporter/pkg"
	"github.com/emylincon/aws_quota_exporter/pkg/scrape"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ServiceConfig struct
type ServiceConfig struct {
	region      string
	profile     string
	serviceCode string
}

func main() {
	var (
		configFile = flag.String("config.file", "config.yaml", "Path to configuration file")
		promPort   = flag.Int("prom.port", 9150, "port to expose prometheus metrics")
	)
	flag.Parse()
	version := "0.0.0"
	// Make Prometheus client aware of our collectors.
	fmt.Println("Config file:", *configFile)
	sc := ServiceConfig{region: "us-east-1", profile: "emeka", serviceCode: "lambda"}
	s, err := scrape.NewScraper(sc.profile)
	if err != nil {
		fmt.Println("Error creating scrape:", err)
		return
	}

	reg := prometheus.NewRegistry()
	qc := pkg.NewPrometheusCollector(s.CreateScraper(sc.region, sc.serviceCode))
	reg.MustRegister(qc)

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
