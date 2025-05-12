package pkg

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	cw "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	sqTypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
)

type MockCloudWatchClient struct {
	CloudWatchClient
}

func (m *MockCloudWatchClient) GetMetricStatistics(ctx context.Context, params *cw.GetMetricStatisticsInput, optFns ...func(*cw.Options)) (*cw.GetMetricStatisticsOutput, error) {
	return &cw.GetMetricStatisticsOutput{
		Datapoints: []cwTypes.Datapoint{
			{
				Average: aws.Float64(50),
			},
		},
	}, nil
}

func TestNewScraper(t *testing.T) {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	tests := []struct {
		name    string
		want    *Scraper
		wantErr bool
	}{
		{
			name: "test New Scraper",
			want: &Scraper{
				cfg: cfg,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewScraper()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewScraper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got.cfg.Region, tt.want.cfg.Region) {
				t.Errorf("NewScraper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScraper_CreateScraper(t *testing.T) {
	type fields struct {
		cfg aws.Config
	}
	type args struct {
		job                 JobConfig
		cacheExpiryDuration time.Duration
		collectUsage        bool
		serveStale          bool
	}
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	failedServiceQuota := errors.New("operation error Service Quotas: ListServiceQuotas, failed to sign request: failed to retrieve credentials: failed to refresh cached credentials, no EC2 IMDS role found, operation error ec2imds: GetMetadata, request canceled, context deadline exceeded")
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func() ([]*PrometheusMetric, error)
		wantErr bool
	}{
		{
			name: "test create scrapper",
			fields: fields{
				cfg: cfg,
			},
			args: args{
				job:                 JobConfig{Regions: []string{"us-west-2"}, ServiceCode: "lambda"},
				cacheExpiryDuration: time.Duration(1) * time.Hour,
			},
			want: func() ([]*PrometheusMetric, error) {
				return nil, failedServiceQuota
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scraper{
				cfg: tt.fields.cfg,
			}

			got := s.CreateScraper(tt.args.job, &tt.args.cacheExpiryDuration, tt.args.serveStale, tt.args.collectUsage)
			d, derr := got()
			r, terr := tt.want()
			if (derr != nil) != tt.wantErr {
				t.Errorf("Scraper.CreateScraper() error = %v, wantErr %v", derr, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(d, r) {
				t.Errorf("Scraper.CreateScraper() = %v:%v, want %v:%v", d, derr, r, terr)
			}
		})
	}
}

func Test_validateRoleARN(t *testing.T) {
	type args struct {
		role string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Valid role ARN",
			args: args{role: "arn:aws:iam::012345678901:role/aws-quota-exporter"},
			want: true,
		},
		{
			name: "Invalid role ARN",
			args: args{role: "arn:aws:iam::012345678901:user/aws-quota-exporter"},
			want: false,
		},
		{
			name: "Not an ARN",
			args: args{role: "foo"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateRoleARN(tt.args.role); got != tt.want {
				t.Errorf("validateRoleARN() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_getQuotasUsage(t *testing.T) {
	type args struct {
		ctx    context.Context
		quotas []sqTypes.ServiceQuota
		region string
	}
	tests := []struct {
		name string
		args args
		want []QuotaUsage
	}{
		{
			name: "Test with empty quotas",
			args: args{
				ctx:    context.TODO(),
				quotas: []sqTypes.ServiceQuota{},
				region: "us-west-2",
			},
			want: []QuotaUsage{},
		},
		{
			name: "Test with quotas without usage metrics",
			args: args{
				ctx: context.TODO(),
				quotas: []sqTypes.ServiceQuota{
					{
						QuotaCode: aws.String("L-12345"),
						QuotaName: aws.String("Test Quota"),
						Value:     aws.Float64(100),
					},
				},
				region: "us-west-2",
			},
			want: []QuotaUsage{
				{
					Quota: sqTypes.ServiceQuota{
						QuotaCode: aws.String("L-12345"),
						QuotaName: aws.String("Test Quota"),
						Value:     aws.Float64(100),
					},
					Usage: 0,
				},
			},
		},
		{
			name: "Test with quotas with usage metrics",
			args: args{
				ctx: context.TODO(),
				quotas: []sqTypes.ServiceQuota{
					{
						ServiceCode: aws.String("test"),
						QuotaCode:   aws.String("L-12345"),
						QuotaName:   aws.String("Test Quota"),
						Value:       aws.Float64(100),
						UsageMetric: &sqTypes.MetricInfo{
							MetricName:                    aws.String("CPUUtilization"),
							MetricNamespace:               aws.String("AWS/EC2"),
							MetricDimensions:              map[string]string{"InstanceId": "i-1234567890abcdef0"},
							MetricStatisticRecommendation: aws.String("Average"),
						},
					},
				},
				region: "us-west-2",
			},
			want: []QuotaUsage{
				{
					Quota: sqTypes.ServiceQuota{
						ServiceCode: aws.String("test"),
						QuotaCode:   aws.String("L-12345"),
						QuotaName:   aws.String("Test Quota"),
						Value:       aws.Float64(100),
						UsageMetric: &sqTypes.MetricInfo{
							MetricName:                    aws.String("CPUUtilization"),
							MetricNamespace:               aws.String("AWS/EC2"),
							MetricDimensions:              map[string]string{"InstanceId": "i-1234567890abcdef0"},
							MetricStatisticRecommendation: aws.String("Average"),
						},
					},
					Usage: 50, // This should be set to the actual usage value from the mock CloudWatch client
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCWClient := &MockCloudWatchClient{}
			got := getQuotasUsage(tt.args.ctx, tt.args.quotas, mockCWClient, tt.args.region)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getQuotasUsage() = %v, want %v", got, tt.want)
			}
		})
	}
}
