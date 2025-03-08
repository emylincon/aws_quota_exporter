# aqe

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.3](https://img.shields.io/badge/AppVersion-0.1.3-informational?style=flat-square)

A Helm chart for aws quota exporter

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://charts/grafana | grafana | 6.58.4 |
| https://prometheus-community.github.io/helm-charts | prometheus | 23.1.0 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `100` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| configmap."config.yml" | string | `"jobs:\n  - serviceCode: lambda\n    regions:\n      - us-west-2\n      - us-east-2\n  - serviceCode: cloudformation\n    regions:\n      - us-west-2\n      - us-east-2\n  - serviceCode: ec2\n    regions:\n      - us-west-2\n      - us-east-2\n"` |  |
| env.AWS_REGION | string | `"us-west-2"` |  |
| fullnameOverride | string | `""` |  |
| grafana."grafana.ini"."auth.anonymous".enabled | bool | `true` |  |
| grafana."grafana.ini"."auth.basic".enabled | bool | `false` |  |
| grafana."grafana.ini"."auth.grafana_com".enabled | bool | `false` |  |
| grafana."grafana.ini".hide_version | bool | `true` |  |
| grafana."grafana.ini".org_name | bool | `true` |  |
| grafana."grafana.ini".org_role | string | `"Admin"` |  |
| grafana.dashboardProviders."dashboardproviders.yaml".apiVersion | int | `1` |  |
| grafana.dashboardProviders."dashboardproviders.yaml".providers[0].disableDeletion | bool | `false` |  |
| grafana.dashboardProviders."dashboardproviders.yaml".providers[0].editable | bool | `true` |  |
| grafana.dashboardProviders."dashboardproviders.yaml".providers[0].folder | string | `""` |  |
| grafana.dashboardProviders."dashboardproviders.yaml".providers[0].name | string | `"prometheus"` |  |
| grafana.dashboardProviders."dashboardproviders.yaml".providers[0].options.path | string | `"/var/lib/grafana/dashboards/prometheus"` |  |
| grafana.dashboardProviders."dashboardproviders.yaml".providers[0].orgId | int | `1` |  |
| grafana.dashboardProviders."dashboardproviders.yaml".providers[0].type | string | `"file"` |  |
| grafana.dashboards.prometheus.quotas.file | string | `"dashboards/quotas.json"` |  |
| grafana.datasources."datasources.yaml".apiVersion | int | `1` |  |
| grafana.datasources."datasources.yaml".datasources[0].access | string | `"proxy"` |  |
| grafana.datasources."datasources.yaml".datasources[0].editable | bool | `true` |  |
| grafana.datasources."datasources.yaml".datasources[0].name | string | `"prometheus"` |  |
| grafana.datasources."datasources.yaml".datasources[0].orgId | int | `1` |  |
| grafana.datasources."datasources.yaml".datasources[0].type | string | `"prometheus"` |  |
| grafana.datasources."datasources.yaml".datasources[0].url | string | `"http://aqe-prometheus-server.default.svc.cluster.local:80"` |  |
| grafana.enabled | bool | `true` |  |
| grafana.image.tag | string | `"11.5.2"` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"ugwuanyi/aqe"` |  |
| image.tag | string | `"latest"` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations | object | `{}` |  |
| ingress.className | string | `""` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts[0].host | string | `"aqe.chart.emylincon.com"` |  |
| ingress.hosts[0].paths[0].path | string | `"/"` |  |
| ingress.hosts[0].paths[0].pathType | string | `"Prefix"` |  |
| ingress.tls | list | `[]` |  |
| livenessProbe.httpGet.path | string | `"/"` |  |
| livenessProbe.httpGet.port | string | `"http"` |  |
| livenessProbe.initialDelaySeconds | int | `60` |  |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podArgs[0] | string | `"--log.level=info"` |  |
| podArgs[1] | string | `"--log.format=text"` |  |
| podSecurityContext | object | `{}` |  |
| prometheus.alertmanager.enabled | bool | `false` |  |
| prometheus.enabled | bool | `true` |  |
| prometheus.kube-state-metrics.enabled | bool | `false` |  |
| prometheus.prometheus-node-exporter.enabled | bool | `false` |  |
| prometheus.prometheus-pushgateway.enabled | bool | `false` |  |
| prometheus.serverFiles."prometheus.yml".scrape_configs[0].job_name | string | `"prometheus"` |  |
| prometheus.serverFiles."prometheus.yml".scrape_configs[0].scrape_interval | string | `"5s"` |  |
| prometheus.serverFiles."prometheus.yml".scrape_configs[0].static_configs[0].targets[0] | string | `"localhost:9090"` |  |
| prometheus.serverFiles."prometheus.yml".scrape_configs[1].job_name | string | `"aws_quota_exporter"` |  |
| prometheus.serverFiles."prometheus.yml".scrape_configs[1].scrape_interval | string | `"15s"` |  |
| prometheus.serverFiles."prometheus.yml".scrape_configs[1].static_configs[0].targets[0] | string | `"aqe.default.svc.cluster.local:10100"` |  |
| readinessProbe.httpGet.path | string | `"/"` |  |
| readinessProbe.httpGet.port | string | `"http"` |  |
| readinessProbe.initialDelaySeconds | int | `60` |  |
| replicaCount | int | `1` |  |
| resources | object | `{}` |  |
| secret.AWS_ACCESS_KEY_ID | string | `""` |  |
| secret.AWS_SECRET_ACCESS_KEY | string | `""` |  |
| securityContext | object | `{}` |  |
| service.port | int | `10100` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `false` |  |
| serviceAccount.name | string | `""` |  |
| serviceMonitor.additionalLabels | object | `{}` |  |
| serviceMonitor.create | bool | `false` |  |
| serviceMonitor.interval | string | `nil` |  |
| serviceMonitor.metricRelabelings | list | `[]` |  |
| serviceMonitor.namespace | string | `nil` |  |
| serviceMonitor.relabelings | list | `[]` |  |
| serviceMonitor.scrapeTimeout | string | `nil` |  |
| tolerations | list | `[]` |  |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.11.0](https://github.com/norwoodj/helm-docs/releases/v1.11.0)
