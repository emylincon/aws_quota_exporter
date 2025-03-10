# AWS Quota Exporter Helm Chart

This Helm chart deploys the AWS Quota Exporter on a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.12+
- Helm 3.0+

## Installation

1. Add the Helm repository:

    ```sh
    helm repo add aws-quota-exporter https://emylincon.github.io/aws_quota_exporter
    helm repo update
    ```

2. Install the chart with the release name `aqe`:

    ```sh
    helm install aqe aws-quota-exporter/aqe
    ```

## Uninstallation

To uninstall/delete the `aqe` deployment:

```sh
helm uninstall aqe
```

## Chart dependencies
### Prometheus
The prometheus chart dependency is managed [here](https://github.com/prometheus-community/helm-charts)

### Grafana
The grafana chart dependency is managed [here](https://github.com/grafana/helm-charts/tree/main/charts/grafana)
