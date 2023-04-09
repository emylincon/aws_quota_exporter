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
