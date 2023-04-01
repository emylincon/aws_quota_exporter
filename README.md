# aws_quota_exporter
Export AWS quotas on Prometheus

# Usage
* Run the following command
```
go run . --prom.port=10100 --config.file=config.yml
```
* Example of `config.yml`
```yaml
jobs:
  - serviceCode: lambda
    regions:
      - us-west-1
      - us-east-1
  - serviceCode: cloudformation
    regions:
      - us-west-1
      - us-east-1
```
# AWS Authentication
This program relies on the `AWS SDK for Go V2` for handling authentication.
The AWS SDK uses its default credential chain to find AWS credentials. This default credential chain looks for credentials in the following order:

1. **Environment variables**
    1. **Static Credentials:** `(AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN)`
    2. **Web Identity Token:** `(AWS_WEB_IDENTITY_TOKEN_FILE)`

2. **Shared configuration files**
    * SDK defaults to `credentials file` and `config file` under `.aws` folder that is placed in the home folder on the host.

3. IAM role for tasks.
4. IAM role for Amazon EC2.

*By default, the SDK checks the `AWS_PROFILE` environment variable to determine which profile to use. If no `AWS_PROFILE` variable is set, the SDK uses the default profile.*

*To set profile to use:*
```bash
$ AWS_PROFILE=test_profile
```
## AWS Permission Required
The exporter requires the AWS managed policy `ServiceQuotasReadOnlyAccess`. This also depends on the jobs specified in the `config.yml` file, as all of the permissions are probably not required. The permissions included in `ServiceQuotasReadOnlyAccess` are as follows in policy document:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "autoscaling:DescribeAccountLimits",
                "cloudformation:DescribeAccountLimits",
                "cloudwatch:DescribeAlarmsForMetric",
                "cloudwatch:DescribeAlarms",
                "cloudwatch:GetMetricData",
                "cloudwatch:GetMetricStatistics",
                "dynamodb:DescribeLimits",
                "elasticloadbalancing:DescribeAccountLimits",
                "iam:GetAccountSummary",
                "kinesis:DescribeLimits",
                "organizations:DescribeAccount",
                "organizations:DescribeOrganization",
                "organizations:ListAWSServiceAccessForOrganization",
                "rds:DescribeAccountAttributes",
                "route53:GetAccountLimit",
                "tag:GetTagKeys",
                "tag:GetTagValues",
                "servicequotas:GetAssociationForServiceQuotaTemplate",
                "servicequotas:GetAWSDefaultServiceQuota",
                "servicequotas:GetRequestedServiceQuotaChange",
                "servicequotas:GetServiceQuota",
                "servicequotas:GetServiceQuotaIncreaseRequestFromTemplate",
                "servicequotas:ListAWSDefaultServiceQuotas",
                "servicequotas:ListRequestedServiceQuotaChangeHistory",
                "servicequotas:ListRequestedServiceQuotaChangeHistoryByQuota",
                "servicequotas:ListServices",
                "servicequotas:ListServiceQuotas",
                "servicequotas:ListServiceQuotaIncreaseRequestsInTemplate",
                "servicequotas:ListTagsForResource"
            ],
            "Resource": "*"
        }
    ]
}
```
*Please Remove permissions that you would not use*

# Useful resources
* include default [port](https://github.com/prometheus/prometheus/wiki/Default-port-allocations) here when finished
* [Guide on how to write an exporter](https://prometheus.io/docs/instrumenting/writing_exporters/)
* [AWS Service Quota Documentation](https://docs.aws.amazon.com/general/latest/gr/aws_service_limits.html)
    * [list-service-quotas](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-service-quotas.html): Lists the `applied quota values` for the specified AWS service. For some quotas, only the default values are available. If the applied quota value is not available for a quota, the quota is not retrieved
    * [list-aws-default-service-quotas](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-aws-default-service-quotas.html): Lists the `default values` for the quotas for the specified AWS service. A default value does not reflect any quota increases.

## References
* [yace_exporter](https://github.com/nerdswords/yet-another-cloudwatch-exporter/)
* [basics-exporter](https://github.com/antonputra/tutorials/blob/main/lessons/141/prometheus-nginx-exporter/)
