package pkg

// Documentation for interacting with aws-sdk-go-v2 https://aws.github.io/aws-sdk-go-v2/docs/getting-started/
import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	sq "github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"golang.org/x/exp/slog"
)

// Scraper struct
type Scraper struct {
	cfg aws.Config
}

// NewScraper creates a new Scraper
func NewScraper() (*Scraper, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return &Scraper{}, err
	}

	return &Scraper{cfg: cfg}, nil
}

var maxResults int32 = 100

// CreateScraper Scrape Quotas from AWS
func (s *Scraper) CreateScraper(regions []string, serviceCode string, cacheExpiryDuration time.Duration) func() ([]*PrometheusMetric, error) {
	// create new cache for service
	cacheStore := NewCache(serviceCode+".json", cacheExpiryDuration)

	return func() ([]*PrometheusMetric, error) {
		// logging start metrics collection
		l := slog.With("serviceCode", serviceCode, "regions", regions, logGroup)
		start := time.Now()
		cacheData, err := cacheStore.Read()
		if err == nil {
			l.Info("Metrics Read from cache",
				"duration", time.Since(start),
			)
			return cacheData, nil
		} else if err == ErrCacheExpired {
			l.Debug("Cache Read", "msg", err)
		} else {
			l.Debug("Cache Read Error", "error", err)
		}

		l.Info("Scrapping metrics")

		ctx := context.Background()
		sclient := sq.NewFromConfig(s.cfg)
		input := sq.ListServiceQuotasInput{ServiceCode: &serviceCode, MaxResults: &maxResults}
		metricList := []*PrometheusMetric{}

		for _, region := range regions {
			metrics, err := getServiceQuotas(ctx, region, &input, sclient)
			if err != nil {
				l.ErrorCtx(ctx, "Failed to get service quotas",
					"error", err,
				)
				return nil, err
			}

			metricList = append(metricList, metrics...)
		}
		err = cacheStore.Write(metricList)
		if err != nil {
			l.Debug("Cache Write error", "error", err)
		}
		l.Info("Metrics Scrapped",
			"duration", time.Since(start),
		)
		return metricList, nil

	}

}

// Transform to prometheus format
func Transform(quotas *sq.ListServiceQuotasOutput, defaultQuotas *sq.ListAWSDefaultServiceQuotasOutput, region string) ([]*PrometheusMetric, error) {
	metrics := []*PrometheusMetric{}
	check := map[string]bool{}
	for _, v := range quotas.Quotas {
		metricName := createMetricName(*v.ServiceCode, *v.QuotaName)
		metric := &PrometheusMetric{
			Name:   metricName,
			Value:  *v.Value,
			Labels: map[string]string{"adjustable": strconv.FormatBool(v.Adjustable), "global_quota": strconv.FormatBool(v.GlobalQuota), "unit": *v.Unit, "region": region},
			Desc:   *v.QuotaName,
		}
		metrics = append(metrics, metric)
		check[metricName] = true
	}
	for _, d := range defaultQuotas.Quotas {
		metricName := createMetricName(*d.ServiceCode, *d.QuotaName)
		if _, ok := check[metricName]; !ok {
			metric := &PrometheusMetric{
				Name:   metricName,
				Value:  *d.Value,
				Labels: map[string]string{"adjustable": strconv.FormatBool(d.Adjustable), "global_quota": strconv.FormatBool(d.GlobalQuota), "unit": *d.Unit, "region": region},
				Desc:   *d.QuotaName,
			}
			metrics = append(metrics, metric)
		}
	}
	return metrics, nil
}

func createMetricName(serviceCode, quotaName string) string {
	return fmt.Sprintf("aws_quota_%s_%s", serviceCode, PromString(quotaName))
}

func getServiceQuotas(ctx context.Context, region string, sqInput *sq.ListServiceQuotasInput, client *sq.Client) ([]*PrometheusMetric, error) {
	opts := func(o *sq.Options) { o.Region = region }

	// Get applied Quotas
	r, err := client.ListServiceQuotas(ctx, sqInput, opts)
	if err != nil {
		return nil, err
	}
	for {
		if r.NextToken == nil {
			break
		}
		sqInput.NextToken = r.NextToken
		rn, err := client.ListServiceQuotas(ctx, sqInput, opts)
		if err != nil {
			return nil, err
		}
		r.Quotas = append(r.Quotas, rn.Quotas...)
		r.NextToken = rn.NextToken

	}
	asqInput := &sq.ListAWSDefaultServiceQuotasInput{ServiceCode: sqInput.ServiceCode, MaxResults: &maxResults}
	// Get default Quotas
	d, err := client.ListAWSDefaultServiceQuotas(ctx, asqInput, opts)
	if err != nil {
		return nil, err
	}
	for {
		if d.NextToken == nil {
			break
		}
		asqInput.NextToken = d.NextToken
		dn, err := client.ListServiceQuotas(ctx, sqInput, opts)
		if err != nil {
			return nil, err
		}
		r.Quotas = append(d.Quotas, dn.Quotas...)
		r.NextToken = dn.NextToken

	}
	return Transform(r, d, region)
}
