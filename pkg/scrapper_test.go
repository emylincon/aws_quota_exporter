package pkg

import (
	"context"
	"reflect"
	"testing"

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
