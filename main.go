package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metrics struct {
	size  prometheus.Counter
	value prometheus.Gauge
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		size: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "nginx",
			Name:      "size_bytes_total",
			Help:      "Total bytes sent to the clients.",
		}),
		value: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "nginx",
			Name:      "emeka_value_total",
			Help:      "Emeka test",
		}),
	}
	reg.MustRegister(m.size, m.value)
	return m
}

func main() {
	var (
		configFile     = flag.String("config.file", "config.yaml", "Path to configuration file")
		scrapeInterval = flag.Int("scrape-interval", 300, "Seconds to wait between scraping the AWS metrics")
		promPort       = flag.Int("prom.port", 9150, "port to expose prometheus metrics")
	)
	flag.Parse()
	version := "0.0.0"
	// Make Prometheus client aware of our collectors.
	fmt.Println("Config file:", *configFile)

	reg := prometheus.NewRegistry()

	m := NewMetrics(reg)
	go GenData(m, *scrapeInterval)

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

func GenData(m *metrics, scrapeInterval int) {
	for {

		m.size.Add(10)
		m.value.Set(55)

		time.Sleep(time.Duration(scrapeInterval) * time.Second)

	}
}
