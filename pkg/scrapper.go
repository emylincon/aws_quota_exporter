package pkg

// Documentation for interacting with aws-sdk-go-v2 https://aws.github.io/aws-sdk-go-v2/docs/getting-started/
import (
	"context"
	"fmt"
	"strconv"
	"sync"
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

type chanData struct {
	metrics []*PrometheusMetric
	err     error
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
		c := make(chan chanData)
		// create goroutine workers
		for _, region := range regions {
			go getServiceQuotas(ctx, region, &input, sclient, c)
		}
		// retrieve channel results from goroutines
		for i := 0; i < len(regions); i++ {
			data := <-c
			if data.err != nil {
				l.ErrorCtx(ctx, "Failed to get service quotas",
					"error", data.err,
				)
				return nil, data.err
			}

			metricList = append(metricList, data.metrics...)
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

func getServiceQuotas(ctx context.Context, region string, sqInput *sq.ListServiceQuotasInput, client *sq.Client, c chan chanData) {
	opts := func(o *sq.Options) { o.Region = region }
	asqInput := &sq.ListAWSDefaultServiceQuotasInput{ServiceCode: sqInput.ServiceCode, MaxResults: &maxResults}
	var wg sync.WaitGroup
	var r *sq.ListServiceQuotasOutput
	var d *sq.ListAWSDefaultServiceQuotasOutput
	errs := [2]error{}

	wg.Add(2)

	// Get applied Quotas
	go func() {
		r, errs[0] = getListServiceQuotas(ctx, client, opts, sqInput)
		wg.Done()
	}()

	// Get default Quotas
	go func() {
		d, errs[1] = getDefaultListServiceQuotas(ctx, client, opts, asqInput)
		wg.Done()
	}()

	wg.Wait()
	for _, err := range errs {
		if err != nil {
			data := chanData{
				metrics: nil,
				err:     err,
			}
			c <- data
			return
		}

	}

	m, err := Transform(r, d, region)
	data := chanData{
		metrics: m,
		err:     err,
	}
	c <- data
}

func getListServiceQuotas(ctx context.Context, client *sq.Client, opts func(o *sq.Options), sqInput *sq.ListServiceQuotasInput) (*sq.ListServiceQuotasOutput, error) {

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
	return r, nil
}

func getDefaultListServiceQuotas(ctx context.Context, client *sq.Client, opts func(o *sq.Options), sqInput *sq.ListAWSDefaultServiceQuotasInput) (*sq.ListAWSDefaultServiceQuotasOutput, error) {

	r, err := client.ListAWSDefaultServiceQuotas(ctx, sqInput, opts)
	if err != nil {
		return nil, err
	}
	for {
		if r.NextToken == nil {
			break
		}
		sqInput.NextToken = r.NextToken
		rn, err := client.ListAWSDefaultServiceQuotas(ctx, sqInput, opts)
		if err != nil {
			return nil, err
		}
		r.Quotas = append(r.Quotas, rn.Quotas...)
		r.NextToken = rn.NextToken

	}
	return r, nil
}
