package pkg

import (
	"fmt"
	"math/rand"
	"time"
)

// GetRandomData returns a random data for testing
func GetRandomData() ([]*PrometheusMetric, error) {

	t := rand.New(rand.NewSource(time.Now().UnixNano()))

	r := func(max float64, min float64) float64 { return min + t.Float64()*(max-min) }
	rr := func() float64 { return r(1000, 0) }
	desc := "Help is not implemented yet."
	metrics := []*PrometheusMetric{
		{Name: "lambda_execution_concurrency_avg", Labels: map[string]string{"region": "us-west-2", "account": "1111"}, Value: rr(), Desc: desc},
		{Name: "lambda_execution_concurrency_min", Labels: map[string]string{"region": "us-west-2", "account": "1111"}, Value: rr(), Desc: desc},
		{Name: "lambda_execution_concurrency_max", Labels: map[string]string{"region": "us-west-2", "account": "1111"}, Value: rr(), Desc: desc},
		{Name: "lambda_execution_concurrency_avg", Labels: map[string]string{"region": "us-east-2", "account": "1111"}, Value: rr(), Desc: desc},
		{Name: "lambda_execution_concurrency_min", Labels: map[string]string{"region": "us-east-2", "account": "1111"}, Value: rr(), Desc: desc},
		{Name: "lambda_execution_concurrency_max", Labels: map[string]string{"region": "us-east-2", "account": "1111"}, Value: rr(), Desc: desc},
	}
	fmt.Println("metrics:", metrics)
	return metrics, nil
}
