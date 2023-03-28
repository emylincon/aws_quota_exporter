package scrape

// Documentation for interacting with aws-sdk-go-v2 https://aws.github.io/aws-sdk-go-v2/docs/getting-started/
import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	sq "github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"github.com/emylincon/aws_quota_exporter/pkg"
)

// Scraper struct
type Scraper struct {
	cfg aws.Config
}

// NewScraper creates a new Scraper
func NewScraper(profileName string) (*Scraper, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(profileName))
	if err != nil {
		return &Scraper{}, err
	}

	return &Scraper{cfg: cfg}, nil
}

// CreateScraper Scrape Quotas from AWS
func (s *Scraper) CreateScraper(regions []string, serviceCode string) func() ([]*pkg.PrometheusMetric, error) {
	return func() ([]*pkg.PrometheusMetric, error) {
		ctx := context.Background()

		sclient := sq.NewFromConfig(s.cfg)
		input := sq.ListServiceQuotasInput{ServiceCode: &serviceCode}
		metricList := []*pkg.PrometheusMetric{}

		for _, region := range regions {
			metrics, err := getServiceQuotas(ctx, region, &input, sclient)
			if err != nil {
				fmt.Printf("Failed to get service quotas: %v", err)
				return nil, err // TODO: return errors
			}

			metricList = append(metricList, metrics...)
		}
		return metricList, nil

	}

}

// Transform to prometheus format
func Transform(quotas *sq.ListServiceQuotasOutput, region string) ([]*pkg.PrometheusMetric, error) {
	metrics := make([]*pkg.PrometheusMetric, len(quotas.Quotas))
	for i, v := range quotas.Quotas {
		metric := &pkg.PrometheusMetric{
			Name:   createMetricName(*v.ServiceCode, *v.QuotaName),
			Value:  *v.Value,
			Labels: map[string]string{"adjustable": strconv.FormatBool(v.Adjustable), "global_quota": strconv.FormatBool(v.GlobalQuota), "unit": *v.Unit, "region": region},
			Desc:   *v.QuotaName,
		}
		metrics[i] = metric
	}
	return metrics, nil
}

func createMetricName(serviceCode, quotaName string) string {
	return fmt.Sprintf("aws_quota_%s_%s", serviceCode, pkg.PromString(quotaName))
}

func getServiceQuotas(ctx context.Context, region string, sqInput *sq.ListServiceQuotasInput, client *sq.Client) ([]*pkg.PrometheusMetric, error) {
	opts := func(o *sq.Options) { o.Region = region }

	r, err := client.ListServiceQuotas(ctx, sqInput, opts)
	if err != nil {
		fmt.Printf("Failed to get service quotas: %v", err)
		return nil, err
	}
	return Transform(r, region)
}
