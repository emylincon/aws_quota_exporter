package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthzHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	writer := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
	writer(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	if string(data) != "ok" {
		t.Errorf("expected 'ok' got %v", string(data))
	}
}
func TestBuildInfoMetrics(t *testing.T) {
	metrics, err := buildInfoMetrics()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	metric := metrics[0]
	if metric.Name != "aqe_build_info" {
		t.Errorf("expected metric name 'aqe_build_info', got %s", metric.Name)
	}

	if metric.Value != 1 {
		t.Errorf("expected metric value 1, got %v", metric.Value)
	}

	expectedLabels := []string{"app", "version", "build_date", "platform", "commit", "go_version"}
	for _, label := range expectedLabels {
		if _, exists := metric.Labels[label]; !exists {
			t.Errorf("expected label '%s' to exist in metric labels", label)
		}
	}
}
