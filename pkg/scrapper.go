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
	sq "github.com/aws/aws-sdk-go-v2/service/servicequotas"
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

// NewScraper creates a new Scraper
func NewScraper() (*Scraper, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return &Scraper{}, err
	}

	return &Scraper{cfg: cfg}, nil
}

// CreateScraper Scrape Quotas from AWS
func (s *Scraper) CreateScraper(job JobConfig, cacheDuration *time.Duration) func() ([]*PrometheusMetric, error) {

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
		sclient := sq.NewFromConfig(cfg)
		input := sq.ListServiceQuotasInput{ServiceCode: &job.ServiceCode, MaxResults: &maxResults}
		metricList := []*PrometheusMetric{}
		c := make(chan chanData)
		// create goroutine workers
		for _, region := range job.Regions {
			go getServiceQuotas(ctx, region, AccountID, &input, sclient, c)
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
func Transform(quotas *sq.ListServiceQuotasOutput, defaultQuotas *sq.ListAWSDefaultServiceQuotasOutput, region, account string) ([]*PrometheusMetric, error) {
	quotas.Quotas = append(quotas.Quotas, defaultQuotas.Quotas...)
	g := NewGrouping(maxSimilarity, region, account)
	mg, metrics := g.GroupMetrics(quotas.Quotas)
	for _, d := range mg {
		if len(d) == 1 { // one item in group
			quota := d[0]
			metric := &PrometheusMetric{
				Name:  createMetricName(*quota.Quota.ServiceCode, *quota.Quota.QuotaName),
				Value: *quota.Quota.Value,
				Labels: map[string]string{
					"adjustable":   strconv.FormatBool(quota.Quota.Adjustable),
					"global_quota": strconv.FormatBool(quota.Quota.GlobalQuota),
					"unit":         *quota.Quota.Unit,
					"region":       region,
					"account":      account,
					"name":         *quota.Quota.QuotaName,
				},
				Desc: createDescription(*quota.Quota.ServiceName, *quota.Quota.QuotaName),
			}
			metrics = append(metrics, metric)
		}
	}
	return metrics, nil
}

func createMetricName(serviceCode, quotaName string) string {
	return fmt.Sprintf("aws_quota_%s_%s", serviceCode, PromString(quotaName))
}

func createDescription(serviceName, quotaName string) string {
	return fmt.Sprintf("%s: %s", serviceName, quotaName)
}

func getServiceQuotas(ctx context.Context, region, account string, sqInput *sq.ListServiceQuotasInput, client *sq.Client, c chan chanData) {
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

	m, err := Transform(r, d, region, account)
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
