package pkg

import (
	"reflect"
	"testing"
)

func TestNewQuotaConfig(t *testing.T) {
	type args struct {
		configFile string
	}
	tests := []struct {
		name    string
		args    args
		want    *QuotaConfig
		wantErr bool
	}{
		{
			name: "Check config is loaded correctly",
			args: args{
				configFile: "../docker/aws_quota_exporter/config.yml",
			},
			want: &QuotaConfig{
				Jobs: []JobConfig{
					{
						ServiceCode: "lambda",
						Regions:     []string{"us-west-2", "us-east-2"},
					},
					{
						ServiceCode: "cloudformation",
						Regions:     []string{"us-west-2", "us-east-2"},
					},
					{
						ServiceCode: "ec2",
						Regions:     []string{"us-west-2", "us-east-2"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "wrong config path",
			args: args{
				configFile: "wrong_path/config.yml",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewQuotaConfig(tt.args.configFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewQuotaConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewQuotaConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
