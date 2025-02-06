package pkg

// Documentation for interacting with aws-sdk-go-v2 https://aws.github.io/aws-sdk-go-v2/docs/getting-started/
import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	cw "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	sq "github.com/aws/aws-sdk-go-v2/service/servicequotas"
	sqTypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"golang.org/x/exp/slog"
)

const (
	maxSimilarity = 0.53
)

var (
	maxResults int32 = 100
)

// Scraper struct
type Scraper struct {
	cfg aws.Config
}

type chanData struct {
	metrics []*PrometheusMetric
	err     error
}

// To combine Quota + Usage data
type QuotaUsage struct {
	Quota sqTypes.ServiceQuota
	Usage float64
}

// NewScraper creates a new Scraper
func NewScraper() (*Scraper, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return &Scraper{}, err
	}

	return &Scraper{cfg: cfg}, nil
}

// CreateScraper Scrape Quotas from AWS
func (s *Scraper) CreateScraper(job JobConfig, cacheDuration *time.Duration, collectUsage bool) func() ([]*PrometheusMetric, error) {

	cfg := s.getAWSConfig(job.Role)
	AccountID := getAWSAccountID(cfg)

	// create new cache for service
	cacheStore, err := NewCache(job.ServiceCode, *cacheDuration)
	if err != nil {
		slog.Warn(fmt.Sprintf("Cache disabled for %s (account %s)", job.ServiceCode, AccountID))
	}

	return func() ([]*PrometheusMetric, error) {
		// logging start metrics collection
		l := slog.With("serviceCode", job.ServiceCode, "regions", job.Regions, logGroup)
		start := time.Now()

		if cacheStore != nil {
			cacheData, err := cacheStore.Read()
			if err == nil {
				l.Debug("Metrics Read from cache",
					"duration", time.Since(start),
				)
				return cacheData, nil
			} else if (err == ErrCacheExpired) || (err == ErrCacheEmpty) {
				l.Info("Cache Read", "msg", err)
			} else {
				l.Info("Cache Read Error", "error", err)
			}
		}

		l.Info("Scrapping metrics")

		ctx := context.Background()
		cfg := s.getAWSConfig(job.Role) // get credentials incase it expires
		sqclient := sq.NewFromConfig(cfg)
		cwclient := cw.NewFromConfig(cfg)
		input := sq.ListServiceQuotasInput{ServiceCode: &job.ServiceCode, MaxResults: &maxResults}
		metricList := []*PrometheusMetric{}
		c := make(chan chanData)
		// create goroutine workers
		for _, region := range job.Regions {
			go getServiceQuotas(ctx, collectUsage, region, AccountID, &input, sqclient, cwclient, c)
		}
		// retrieve channel results from goroutines
		for i := 0; i < len(job.Regions); i++ {
			data := <-c
			if data.err != nil {
				l.ErrorCtx(ctx, "Failed to get service quotas",
					"error", data.err,
				)
				return nil, data.err
			}

			metricList = append(metricList, data.metrics...)
		}

		if cacheStore != nil {
			err = cacheStore.Write(metricList)
			if err != nil {
				l.Debug("Cache Write error", "error", err)
			}
		}

		l.Info("Metrics Scrapped",
			"duration", time.Since(start),
		)
		return metricList, nil

	}

}

func getAWSAccountID(cfg aws.Config) string {
	opts := sts.Options{
		APIOptions:   cfg.APIOptions,
		Region:       cfg.Region,
		Credentials:  cfg.Credentials,
		DefaultsMode: cfg.DefaultsMode,
	}

	stssvc := sts.New(opts)
	input := &sts.GetCallerIdentityInput{}
	ctx := context.Background()
	caller, err := stssvc.GetCallerIdentity(ctx, input)

	if err != nil {
		slog.WarnCtx(ctx, "Failed to get caller identity", "error", err)
		return ""
	}

	return *caller.Account

}

func (s *Scraper) getAWSConfig(role string) aws.Config {
	if role == "" {
		return s.cfg
	}
	if !validateRoleARN(role) {
		slog.Error("Role ARN is not valid", "RoleARN", role)
		os.Exit(1)
	}
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		slog.ErrorCtx(ctx, "Error loading default AWS config", "error", err)
		os.Exit(1)
	}
	// Create the credentials from AssumeRoleProvider to assume the role
	// referenced by the "myRoleARN" ARN.
	stsSvc := sts.NewFromConfig(cfg)
	creds := stscreds.NewAssumeRoleProvider(stsSvc, role)
	cfg.Credentials = aws.NewCredentialsCache(creds)
	return cfg
}

func validateRoleARN(role string) bool {
	if arn.IsARN(role) {
		arnObj, err := arn.Parse(role)
		if err != nil {
			return false
		}
		if strings.HasPrefix(arnObj.Resource, "role/") {
			return true
		}
	}
	return false
}

// Transform to prometheus format
func Transform(quotas []QuotaUsage, collectUsage bool, region, account string) ([]*PrometheusMetric, error) {
	g := NewGrouping(maxSimilarity, region, account)
	mg, metrics := g.GroupMetrics(quotas, collectUsage)
	for _, d := range mg {
		if len(d) == 1 { // one item in group
			var quotaMetric, quotaUsage *PrometheusMetric
			quota := d[0]
			quotaMetric = createPromMetric(quota, "quota", region, account)
			metrics = append(metrics, quotaMetric)

			// if Quota has UsageMetric, create also _usage metric
			if collectUsage && quota.Quota.UsageMetric != nil {
				quotaUsage = createPromMetric(quota, "usage", region, account)
				metrics = append(metrics, quotaUsage)
			}
		}
	}
	return metrics, nil
}

// creates a Prometheus metric based on the given metric group (for single quotas).
func createPromMetric(m MetricGroup, metric_type, region, account string) *PrometheusMetric {
	value := *m.Quota.Value
	if metric_type == "usage" {
		value = m.Usage
	}
	return &PrometheusMetric{
		Name:  createMetricName(*m.Quota.ServiceCode, *m.Quota.QuotaName),
		Value: value,
		Labels: map[string]string{
			"type":         metric_type,
			"adjustable":   strconv.FormatBool(m.Quota.Adjustable),
			"global_quota": strconv.FormatBool(m.Quota.GlobalQuota),
			"unit":         *m.Quota.Unit,
			"region":       region,
			"account":      account,
			"name":         *m.Quota.QuotaName,
		},
		Desc: createDescription(*m.Quota.ServiceName, *m.Quota.QuotaName),
	}
}

func createMetricName(serviceCode, quotaName string) string {
	return fmt.Sprintf("aws_quota_%s_%s", serviceCode, PromString(quotaName))
}

func createDescription(serviceName, quotaName string) string {
	return fmt.Sprintf("%s: %s", serviceName, quotaName)
}

func getServiceQuotas(ctx context.Context, collectUsage bool, region, account string, sqInput *sq.ListServiceQuotasInput, sqclient *sq.Client, cwclient *cw.Client, c chan chanData) {
	sqOpts := func(o *sq.Options) { o.Region = region }
	asqInput := &sq.ListAWSDefaultServiceQuotasInput{ServiceCode: sqInput.ServiceCode, MaxResults: &maxResults}
	var wg sync.WaitGroup
	var r *sq.ListServiceQuotasOutput
	var d *sq.ListAWSDefaultServiceQuotasOutput
	var quotasUsage []QuotaUsage
	check := map[string]bool{}
	errs := [2]error{}

	wg.Add(2)

	// Get applied Quotas
	go func() {
		r, errs[0] = getListServiceQuotas(ctx, sqclient, sqOpts, sqInput)
		wg.Done()
	}()

	// Get default Quotas
	go func() {
		d, errs[1] = getDefaultListServiceQuotas(ctx, sqclient, sqOpts, asqInput)
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

	// merge applied Quotas with defaults
	quotasMerged := append(r.Quotas, d.Quotas...)
	// loop over all Quotas and fetch CloudWatch usage once
	cwOpts := func(o *cw.Options) { o.Region = region }
	for _, q := range quotasMerged {
		mq := QuotaUsage{q, 0}
		if collectUsage && q.UsageMetric != nil && !check[*q.QuotaCode] {
			var dimensions []cwTypes.Dimension
			for k, v := range q.UsageMetric.MetricDimensions { // form Dimensions filter for GetMetricStatisticsInput based on UsageMetric.MetricDimensions
				dimensions = append(dimensions, cwTypes.Dimension{Name: aws.String(k), Value: aws.String(v)})
			}
			params := &cw.GetMetricStatisticsInput{
				MetricName: aws.String(*q.UsageMetric.MetricName),
				Namespace:  aws.String(*q.UsageMetric.MetricNamespace),
				StartTime:  aws.Time(time.Now().Add(time.Minute * -5)),
				EndTime:    aws.Time(time.Now()),
				Period:     aws.Int32(60 * 5),
				Dimensions: dimensions,
				Statistics: []cwTypes.Statistic{cwTypes.Statistic(*q.UsageMetric.MetricStatisticRecommendation)},
			}
			resp, err := cwclient.GetMetricStatistics(ctx, params, cwOpts)

			if err == nil {
				if len(resp.Datapoints) != 0 { // if Quota has Usage, it will be set, otherwise it's = 0
					mq.Usage = *resp.Datapoints[0].Maximum
				}
			} else {
				slog.Warn("Unable to retrieve CloudWatch usage", "error", err)
			}
			check[*q.QuotaCode] = true
		}
		quotasUsage = append(quotasUsage, mq)
	}

	m, err := Transform(quotasUsage, collectUsage, region, account)
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
	for r.NextToken != nil {
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
	for r.NextToken != nil {
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
