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

## Useful resources
* include default [port](https://github.com/prometheus/prometheus/wiki/Default-port-allocations) here when finished
* [Guide on how to write an exporter](https://prometheus.io/docs/instrumenting/writing_exporters/)
* [AWS Service Quota Documentation](https://docs.aws.amazon.com/general/latest/gr/aws_service_limits.html)
    * [list-service-quotas](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-service-quotas.html): Lists the `applied quota values` for the specified AWS service. For some quotas, only the default values are available. If the applied quota value is not available for a quota, the quota is not retrieved
    * [list-aws-default-service-quotas](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-aws-default-service-quotas.html): Lists the `default values` for the quotas for the specified AWS service. A default value does not reflect any quota increases.

## References
* [yace_exporter](https://github.com/nerdswords/yet-another-cloudwatch-exporter/)
* [basics-exporter](https://github.com/antonputra/tutorials/blob/main/lessons/141/prometheus-nginx-exporter/)
