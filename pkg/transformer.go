package pkg

import "regexp"

// transformer combines a dimension-less metric name and a regular expression pattern.
//
// The regex is used to extract dimensions from an AWS quota name whereas the name
// is generalised to exclude dimensions and used as the metric name.
type transformer struct {
	name string
	re   *regexp.Regexp
}

// transformers is a map of service codes to defined transformer structs.
//
// This allows for reducing the number of regexes parsed when transforming a metric
// by limiting the search space to the quotas service.
// The names of capture groups in the regex are used as label names.
var transformers = map[string][]transformer{
	"ec2": {
		{
			name: "Running Dedicated Hosts",
			re:   regexp.MustCompile(`^Running Dedicated (?P<instance_family>[\w\s,-]+) Hosts$`),
		},
		{
			name: "Running On-Demand instances",
			re:   regexp.MustCompile(`^Running On-Demand (?P<instance_class>[\(\w\s,\)-]+) instances$`),
		},
		{
			name: "All Spot Instance Requests",
			re:   regexp.MustCompile(`^All (?P<instance_class>[\(\w\s,\)-]+) Spot Instance Requests$`),
		},
	},
	"cloudtrail": {
		{
			name: "Transactions per second",
			re:   regexp.MustCompile(`^Transactions per second \(TPS\) for (the )?(?P<api>[\w\s,]+) API(s)?$`),
		},
	},
	"ebs": {
		{
			name: "Concurrent snapshots per volume",
			re:   regexp.MustCompile(`^Concurrent snapshots per (?P<volume_type_name>[\w\s]+) \((?P<volume_type>[\w]+)\) volume$`),
		},
		{
			name: "IOPS for Provisioned IOPS SSD volumes",
			re:   regexp.MustCompile(`^IOPS for Provisioned IOPS SSD \((?P<volume_type>[\w]+)\) volumes$`),
		},
		{
			name: "IOPS modifications for Provisioned IOPS SSD volumes",
			re:   regexp.MustCompile(`^IOPS modifications for Provisioned IOPS SSD \((?P<volume_type>[\w]+)\) volumes$`),
		},
		{
			name: "Storage for volumes in TiB",
			re:   regexp.MustCompile(`^Storage for (?P<volume_type_name>[\w\s]+) \((?P<volume_type>[\w]+)\) volumes, in TiB$`),
		},
		{
			name: "Storage modifications for volumes in TiB",
			re:   regexp.MustCompile(`^Storage modifications for (?P<volume_type_name>[\w\s]+) \((?P<volume_type>[\w]+)\) volumes, in TiB$`),
		},
	},
	"ecr": {
		{
			name: "Rate of requests",
			re:   regexp.MustCompile(`^Rate of (?P<request_type>[\w]+) requests$`),
		},
	},
	"elasticloadbalancing": {
		{
			name: "Listeners per load balancer type",
			re:   regexp.MustCompile(`^Listeners per (?P<type>\w+) Load Balancer$`),
		},
	},
	"kms": {
		{
			name: "Cryptographic operation request rate",
			re:   regexp.MustCompile(`^Cryptographic operations \((?P<key_type>\w+)\) request rate$`),
		},
		{
			name: "GenerateDataKeyPair request rate",
			re:   regexp.MustCompile(`^GenerateDataKeyPair \((?P<key_spec>\w+)\) request rate$`),
		},
		{
			name: "Request rate",
			re:   regexp.MustCompile(`^(?P<operation>\w+) request rate$`),
		},
	},
	"logs": {
		{
			name: "Throttle limit in transactions per second",
			re:   regexp.MustCompile(`^(?P<operation>\w+) throttle limit in transactions per second$`),
		},
	},
	"sagemaker": {
		{
			name: "Endpoint usage",
			re:   regexp.MustCompile(`^(?P<instance_type>[\w\.]+) for endpoint usage$`),
		},
		{
			name: "Notebook instance usage",
			re:   regexp.MustCompile(`^(?P<instance_type>[\w\.]+) for notebook instance usage$`),
		},
		{
			name: "Processing job usage",
			re:   regexp.MustCompile(`^(?P<instance_type>[\w\.]+) for processing job usage$`),
		},
		{
			name: "Spot training job usage",
			re:   regexp.MustCompile(`^(?P<instance_type>[\w\.]+) for spot training job usage$`),
		},
		{
			name: "Training job usage",
			re:   regexp.MustCompile(`^(?P<instance_type>[\w\.]+) for training job usage$`),
		},
		{
			name: "Training warm pool usage",
			re:   regexp.MustCompile(`^(?P<instance_type>[\w\.]+) for training warm pool usage$`),
		},
		{
			name: "Transform job usage",
			re:   regexp.MustCompile(`^(?P<instance_type>[\w\.]+) for transform job usage$`),
		},
		{
			name: "Rate of requests",
			re:   regexp.MustCompile(`^Rate of (?P<operation>\w+) requests$`),
		},
		{
			name: "Apps running",
			re:   regexp.MustCompile(`^(?P<app_type>[\w\s]+) running on (?P<instance_type>[\w.]+) instances?$`),
		},
	},
	"servicequotas": {
		{
			name: "Throttle rate",
			re:   regexp.MustCompile(`^Throttle rate for (?P<operation>[\w]+)$`),
		},
	},
}
