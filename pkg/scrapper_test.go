package pkg

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

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

			got := s.CreateScraper(tt.args.job, &tt.args.cacheExpiryDuration)
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
