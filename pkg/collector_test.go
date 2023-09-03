package pkg

import (
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func Test_createDesc(t *testing.T) {
	type args struct {
		metric *PrometheusMetric
	}
	metric := &PrometheusMetric{
		Name:   "test",
		Labels: map[string]string{"region": "us-east-1"},
		Value:  50,
		Desc:   "test description",
	}
	tests := []struct {
		name string
		args args
		want *prometheus.Desc
	}{
		{
			name: "Describe",
			args: args{
				metric: metric,
			},
			want: prometheus.NewDesc(
				metric.Name,
				metric.Desc,
				nil,
				metric.Labels,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createDesc(tt.args.metric); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createDesc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitize(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"allowed_only_09AZ", "allowed_only_09AZ"},
		{"!@#$%^&*()'’", "____________"},
		{"CamelCaseAllowedOnly", "CamelCaseAllowedOnly"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := sanitize(tc.input)
			if result != tc.expected {
				t.Errorf("Expected: %s, Got: %s", tc.expected, result)
			}
		})
	}
}

func TestPromString(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"SomeText@Here123", "some_text_here123"},
		{"!@#$%^&*()'’", "____________"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := PromString(tc.input)
			if result != tc.expected {
				t.Errorf("Expected: %s, Got: %s", tc.expected, result)
			}
		})
	}
}
