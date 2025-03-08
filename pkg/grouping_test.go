package pkg

import (
	"fmt"
	"testing"

	sqTypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/aws/aws-sdk-go/aws/awsutil"
)

func TestGroupMetrics(t *testing.T) {

	tests := []struct {
		name         string
		groupsLength int
		promLength   int
		collectUsage bool
		quotas       []QuotaUsage
	}{
		{
			name:         "common with only 2 words",
			groupsLength: 3,
			promLength:   0,
			quotas: []QuotaUsage{
				{Quota: sqTypes.ServiceQuota{
					ServiceCode: Ptr("ec2"),
					ServiceName: Ptr("Amazon Elastic Compute Cloud"),
					QuotaName:   Ptr("Test Quota 1"),
					Value:       Ptr(100.0),
					Adjustable:  true,
					GlobalQuota: false,
					Unit:        Ptr("Count"),
					UsageMetric: nil,
					Period:      nil,
					ErrorReason: nil,
				}},
				{Quota: sqTypes.ServiceQuota{
					ServiceCode: Ptr("ec2"),
					ServiceName: Ptr("Amazon Elastic Compute Cloud"),
					QuotaName:   Ptr("Test Quota 2"),
					Value:       Ptr(100.0),
					Adjustable:  true,
					GlobalQuota: false,
					Unit:        Ptr("Count"),
					UsageMetric: nil,
					Period:      nil,
					ErrorReason: nil,
				}},
				{Quota: sqTypes.ServiceQuota{
					ServiceCode: Ptr("ec2"),
					ServiceName: Ptr("Amazon Elastic Compute Cloud"),
					QuotaName:   Ptr("Test Quota 3"),
					Value:       Ptr(100.0),
					Adjustable:  true,
					GlobalQuota: false,
					Unit:        Ptr("Count"),
					UsageMetric: nil,
					Period:      nil,
					ErrorReason: nil,
				}},
			},
		},
		{
			name:         "common with more than 2 words",
			groupsLength: 1,
			promLength:   3,
			quotas: []QuotaUsage{
				{Quota: sqTypes.ServiceQuota{
					ServiceCode: Ptr("ec2"),
					ServiceName: Ptr("Amazon Elastic Compute Cloud"),
					QuotaName:   Ptr("All DL Spot Instance Requests"),
					Value:       Ptr(100.0),
					Adjustable:  true,
					GlobalQuota: false,
					Unit:        Ptr("Count"),
					UsageMetric: nil,
					Period:      nil,
					ErrorReason: nil,
				}},
				{Quota: sqTypes.ServiceQuota{
					ServiceCode: Ptr("ec2"),
					ServiceName: Ptr("Amazon Elastic Compute Cloud"),
					QuotaName:   Ptr("All F Spot Instance Requests"),
					Value:       Ptr(100.0),
					Adjustable:  true,
					GlobalQuota: false,
					Unit:        Ptr("Count"),
					UsageMetric: nil,
					Period:      nil,
					ErrorReason: nil,
				}},
				{Quota: sqTypes.ServiceQuota{
					ServiceCode: Ptr("ec2"),
					ServiceName: Ptr("Amazon Elastic Compute Cloud"),
					QuotaName:   Ptr("All Standard (A, C, D, H, I, M, R, T, Z) Spot Instance Requests"),
					Value:       Ptr(100.0),
					Adjustable:  true,
					GlobalQuota: false,
					Unit:        Ptr("Count"),
					UsageMetric: nil,
					Period:      nil,
					ErrorReason: nil,
				}},
			},
		},
	}

	maxSimilarity := 0.5
	region := "us-west-2"
	account := "123456789012"

	grouping := NewGrouping(maxSimilarity, region, account)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups, promMetrics := grouping.GroupMetrics(tt.quotas, tt.collectUsage)
			if len(groups) != tt.groupsLength {
				fmt.Println("groups=", awsutil.Prettify(groups), "prom", awsutil.Prettify(promMetrics))
				t.Errorf("Expected %d groups, got %d", tt.groupsLength, len(groups))
			}

			if len(promMetrics) != tt.promLength {
				fmt.Println("prom=", awsutil.Prettify(promMetrics))
				t.Errorf("Expected %d Prometheus metrics, got %d", tt.promLength, len(promMetrics))
			}
		})
	}

}

type Any interface {
	string | float64 | bool
}

// Ptr returns a pointer of given parameter
func Ptr[T Any](any T) *T {
	return &any
}
