// Package pkg grouping is used for grouping AWS service quotas and creating Prometheus metrics based on the grouped quotas.
package pkg

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/adrg/strutil/metrics"
	sqTypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/emylincon/golist"
	"golang.org/x/exp/slog"
)

// MetricGroup represents a group of metrics.
type MetricGroup struct {
	Label  string               `json:"label,omitempty"`  // Label for the metric group.
	Common string               `json:"common,omitempty"` // Common part of the metric name.
	Sim    float64              `json:"sim,omitempty"`    // Similarity score.
	Quota  sqTypes.ServiceQuota `json:"quota,omitempty"`  // AWS service quota.
	Usage  float64              `json:"usage,omitempty"`  // AWS service quota usage.
}

// Grouping represents a Grouping instance.
type Grouping struct {
	maxSimilarity float64        // Maximum similarity score.
	region        string         // AWS region.
	account       string         // AWS account.
	accountName   string         // AWS account name.
	repl          *regexp.Regexp // Regular expression for replacing patterns.
}

// NewGrouping initializes a new grouping instance.
func NewGrouping(maxSimilarity float64, jobRegionCfg JobRegion) *Grouping {
	return &Grouping{
		maxSimilarity: maxSimilarity,
		region:        jobRegionCfg.Region,
		account:       jobRegionCfg.AccountID,
		accountName:   jobRegionCfg.AccountName,
		repl:          regexp.MustCompile(` \(.*\)`),
	}
}

// diff computes the difference between two strings.
func (g *Grouping) diff(a, b string) string {
	list := golist.NewList(strings.Split(a, " "))
	other := golist.NewList(strings.Split(b, " "))
	difference, err := list.Difference(other)
	if err != nil {
		slog.Debug("could not diff strings", "error", err)
		return ""
	}
	result, _ := difference.ConvertToSliceString()

	return strings.Join(result, " ")
}

// common computes the common parts between two strings.
func (g *Grouping) common(a, b string) string {
	result := []string{}
	list := strings.Split(a, " ")
	other := golist.NewList(strings.Split(b, " "))
	if list[0] != other.Get(0) {
		return ""
	}
	for _, item := range list {
		if other.Contains(item) {
			result = append(result, item)
		}
	}
	return strings.Join(result, " ")
}

// createPromMetric creates a Prometheus metric based on the given metric group and type.
func (g *Grouping) createPromMetric(m MetricGroup, metricType string) *PrometheusMetric {
	value := *m.Quota.Value
	if metricType == "usage" {
		value = m.Usage
	}
	return &PrometheusMetric{
		Name:  createMetricName(*m.Quota.ServiceCode, m.Common),
		Value: value,
		Labels: map[string]string{
			"type":         metricType,
			"adjustable":   strconv.FormatBool(m.Quota.Adjustable),
			"global_quota": strconv.FormatBool(m.Quota.GlobalQuota),
			"unit":         *m.Quota.Unit,
			"region":       g.region,
			"account":      g.account,
			"account_name": g.accountName,
			"kind":         m.Label,
			"name":         *m.Quota.QuotaName,
			"quota_code":   *m.Quota.QuotaCode,
			"service_code": *m.Quota.ServiceCode,
		},
		Desc: createDescription(*m.Quota.ServiceName, m.Common),
	}
}

// RemoveBrackets removes brackets from metric names.
func (g *Grouping) RemoveBrackets(str string) string {
	return g.repl.ReplaceAllString(str, "")
}

// GroupMetrics groups AWS service quotas.
func (g *Grouping) GroupMetrics(quotas []QuotaUsage, collectUsage bool) (map[string][]MetricGroup, []*PrometheusMetric) {
	promMetrics := []*PrometheusMetric{}
	hem := metrics.NewLevenshtein()
	check := map[string]bool{}

	response := map[string][]MetricGroup{}

	for _, q := range quotas {
		if _, ok := check[*q.Quota.QuotaName]; ok {
			continue
		}
		check[*q.Quota.QuotaName] = true
		if len(response) == 0 {
			response[*q.Quota.QuotaName] = []MetricGroup{{Quota: q.Quota, Usage: q.Usage}}
		} else {
			selected := false
			for key := range response {
				sim := hem.Compare(g.RemoveBrackets(*q.Quota.QuotaName), g.RemoveBrackets(key))
				if sim >= g.maxSimilarity {
					response[key] = append(response[key], MetricGroup{Quota: q.Quota, Usage: q.Usage})
					if len(response[key]) == 2 {
						commonStr := g.common(*response[key][0].Quota.QuotaName, *response[key][1].Quota.QuotaName)
						if commonStr == "" || len(strings.Split(commonStr, " ")) <= 2 { // if the first words of metric names are not the same or common is only two words then skip
							response[key] = response[key][:len(response[key])-1] // remove added metric
							continue
						}
						for i := 0; i < 2; i++ {
							response[key][i].Label = g.diff(*response[key][i].Quota.QuotaName, *response[key][i^1].Quota.QuotaName)
							response[key][i].Common = commonStr
							response[key][i].Sim = hem.Compare(g.RemoveBrackets(*response[key][i].Quota.QuotaName), g.RemoveBrackets(key))
							promMetrics = append(promMetrics, g.createPromMetric(response[key][i], "quota"))
							if collectUsage && response[key][i].Quota.UsageMetric != nil { // add Usage metric if Quota has UsageMetric
								promMetrics = append(promMetrics, g.createPromMetric(response[key][i], "usage"))
							}
						}
					} else if len(response[key]) > 2 {
						_id := len(response[key]) - 1
						if strings.Split(key, " ")[0] != strings.Split(*response[key][_id].Quota.QuotaName, " ")[0] { // if the first words of metric names are not the same then skip
							continue
						}
						response[key][_id].Label = g.diff(*response[key][_id].Quota.QuotaName, key)
						response[key][_id].Common = response[key][0].Common
						response[key][_id].Sim = sim
						promMetrics = append(promMetrics, g.createPromMetric(response[key][_id], "quota"))
						if collectUsage && response[key][_id].Quota.UsageMetric != nil { // add Usage metric if Quota has UsageMetric
							promMetrics = append(promMetrics, g.createPromMetric(response[key][_id], "usage"))
						}
					}
					selected = true
					break
				}
			}
			if !selected {
				response[*q.Quota.QuotaName] = []MetricGroup{{Quota: q.Quota, Usage: q.Usage}}
			}
		}
	}
	return response, promMetrics
}
