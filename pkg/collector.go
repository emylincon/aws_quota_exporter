package pkg

import (
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
)

var replacer = strings.NewReplacer(
	" ", "_",
	",", "_",
	"\t", "_",
	"/", "_",
	"\\", "_",
	".", "_",
	"-", "_",
	":", "_",
	"=", "_",
	"“", "_",
	"@", "_",
	"<", "_",
	">", "_",
	"%", "_percent",
)
var splitRegexp = regexp.MustCompile(`([a-z0-9])([A-Z])`)

// PrometheusMetric data structure
type PrometheusMetric struct {
	Name   string
	Labels map[string]string
	Value  float64
	Desc   string
}

// PrometheusCollector Data structure
type PrometheusCollector struct {
	mutex      sync.RWMutex
	getMetrics func() ([]*PrometheusMetric, error)
}

// NewPrometheusCollector is PrometheusCollector constructor
func NewPrometheusCollector(getMetrics func() ([]*PrometheusMetric, error)) *PrometheusCollector {
	return &PrometheusCollector{
		getMetrics: getMetrics,
	}
}

// Describe metrics
func (p *PrometheusCollector) Describe(descs chan<- *prometheus.Desc) {
	data, err := p.getMetrics()
	if err != nil {
		descs <- prometheus.NewInvalidDesc(err)
		return
	}
	for _, metric := range removeDuplicatedMetrics(data) {
		descs <- createDesc(metric)
	}
}

// Collect metrics
func (p *PrometheusCollector) Collect(metrics chan<- prometheus.Metric) {
	p.mutex.Lock() // To protect metrics from concurrent collects.
	defer p.mutex.Unlock()
	data, err := p.getMetrics()
	if err != nil {
		desc := prometheus.NewDesc(
			"place_holder_prometheus_collector",
			"Help is not implemented yet",
			[]string{},
			nil,
		)
		metrics <- prometheus.NewInvalidMetric(desc, err)
	}
	for _, metric := range removeDuplicatedMetrics(data) {
		metrics <- createMetric(metric)
	}
}

func createDesc(metric *PrometheusMetric) *prometheus.Desc {
	return prometheus.NewDesc(
		metric.Name,
		metric.Desc,
		nil,
		metric.Labels,
	)
}

func createMetric(metric *PrometheusMetric) prometheus.Metric {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        metric.Name,
		Help:        metric.Desc,
		ConstLabels: metric.Labels,
	})

	gauge.Set(metric.Value)

	return gauge
}

func removeDuplicatedMetrics(metrics []*PrometheusMetric) []*PrometheusMetric {
	keys := make(map[string]bool)
	filteredMetrics := []*PrometheusMetric{}
	for _, metric := range metrics {
		check := metric.Name + combineLabels(metric.Labels)
		if _, value := keys[check]; !value {
			keys[check] = true
			filteredMetrics = append(filteredMetrics, metric)
		}
	}
	return filteredMetrics
}

func combineLabels(labels map[string]string) string {
	var combinedLabels string
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		combinedLabels += PromString(k) + PromString(labels[k])
	}
	return combinedLabels
}

// PromString returns prometheus string representation
func PromString(text string) string {
	text = splitString(text)
	return strings.ToLower(sanitize(text))
}

// PromStringTag checks valid string
func PromStringTag(text string, labelsSnakeCase bool) (bool, string) {
	var s string
	if labelsSnakeCase {
		s = PromString(text)
	} else {
		s = sanitize(text)
	}
	return model.LabelName(s).IsValid(), s
}

func sanitize(text string) string {
	return replacer.Replace(text)
}

func splitString(text string) string {
	return splitRegexp.ReplaceAllString(text, `$1.$2`)
}
