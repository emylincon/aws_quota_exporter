# aqe

![Version: 0.0.2](https://img.shields.io/badge/Version-0.0.2-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.2](https://img.shields.io/badge/AppVersion-0.1.2-informational?style=flat-square)

A Helm chart for aws quota exporter

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://grafana.github.io/helm-charts | grafana | 6.58.4 |
| https://prometheus-community.github.io/helm-charts | prometheus | 23.1.0 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `100` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| configmap."config.yml" | string | `"jobs:\n  - serviceCode: lambda\n    regions:\n      - us-west-1\n      - us-east-1\n  - serviceCode: cloudformation\n    regions:\n      - us-west-1\n      - us-east-1\n"` |  |
| env.AWS_REGION | string | `"us-west-1"` |  |
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
| grafana.dashboards.prometheus.quotas.json | string | `"{\n  \"annotations\": {\n    \"list\": [\n      {\n        \"builtIn\": 1,\n        \"datasource\": {\n          \"type\": \"grafana\",\n          \"uid\": \"-- Grafana --\"\n        },\n        \"enable\": true,\n        \"hide\": true,\n        \"iconColor\": \"rgba(0, 211, 255, 1)\",\n        \"name\": \"Annotations & Alerts\",\n        \"target\": {\n          \"limit\": 100,\n          \"matchAny\": false,\n          \"tags\": [],\n          \"type\": \"dashboard\"\n        },\n        \"type\": \"dashboard\"\n      }\n    ]\n  },\n  \"editable\": true,\n  \"fiscalYearStartMonth\": 0,\n  \"graphTooltip\": 0,\n  \"id\": 1,\n  \"links\": [],\n  \"liveNow\": false,\n  \"panels\": [\n    {\n      \"datasource\": {\n        \"type\": \"prometheus\"\n      },\n      \"fieldConfig\": {\n        \"defaults\": {\n          \"mappings\": [],\n          \"thresholds\": {\n            \"mode\": \"absolute\",\n            \"steps\": [\n              {\n                \"color\": \"green\",\n                \"value\": null\n              },\n              {\n                \"color\": \"red\",\n                \"value\": 200\n              }\n            ]\n          },\n          \"unit\": \"short\"\n        },\n        \"overrides\": []\n      },\n      \"gridPos\": {\n        \"h\": 8,\n        \"w\": 12,\n        \"x\": 0,\n        \"y\": 0\n      },\n      \"id\": 6,\n      \"options\": {\n        \"colorMode\": \"background\",\n        \"graphMode\": \"none\",\n        \"justifyMode\": \"auto\",\n        \"orientation\": \"auto\",\n        \"reduceOptions\": {\n          \"calcs\": [\n            \"lastNotNull\"\n          ],\n          \"fields\": \"\",\n          \"values\": false\n        },\n        \"textMode\": \"auto\"\n      },\n      \"pluginVersion\": \"\",\n      \"targets\": [\n        {\n          \"datasource\": {\n            \"type\": \"prometheus\"\n          },\n          \"editorMode\": \"builder\",\n          \"expr\": \"sum by(region) (aws_quota_cloudformation_module_limit_per_account)\",\n          \"legendFormat\": \"__auto\",\n          \"range\": true,\n          \"refId\": \"A\"\n        }\n      ],\n      \"title\": \"Cloudformation module limit per account\",\n      \"type\": \"stat\"\n    },\n    {\n      \"datasource\": {\n        \"type\": \"prometheus\"\n      },\n      \"fieldConfig\": {\n        \"defaults\": {\n          \"mappings\": [],\n          \"thresholds\": {\n            \"mode\": \"percentage\",\n            \"steps\": [\n              {\n                \"color\": \"green\",\n                \"value\": null\n              },\n              {\n                \"color\": \"orange\",\n                \"value\": 70\n              },\n              {\n                \"color\": \"red\",\n                \"value\": 85\n              }\n            ]\n          }\n        },\n        \"overrides\": []\n      },\n      \"gridPos\": {\n        \"h\": 8,\n        \"w\": 12,\n        \"x\": 12,\n        \"y\": 0\n      },\n      \"id\": 4,\n      \"options\": {\n        \"orientation\": \"auto\",\n        \"reduceOptions\": {\n          \"calcs\": [\n            \"lastNotNull\"\n          ],\n          \"fields\": \"\",\n          \"values\": false\n        },\n        \"showThresholdLabels\": false,\n        \"showThresholdMarkers\": true\n      },\n      \"pluginVersion\": \"\",\n      \"targets\": [\n        {\n          \"datasource\": {\n            \"type\": \"prometheus\"\n          },\n          \"editorMode\": \"builder\",\n          \"expr\": \"sum by(region) (aws_quota_cloudformation_stack_count)\",\n          \"legendFormat\": \"__auto\",\n          \"range\": true,\n          \"refId\": \"A\"\n        }\n      ],\n      \"title\": \"Cloudformation Stack count\",\n      \"type\": \"gauge\"\n    },\n    {\n      \"datasource\": {\n        \"type\": \"prometheus\"\n      },\n      \"fieldConfig\": {\n        \"defaults\": {\n          \"mappings\": [],\n          \"thresholds\": {\n            \"mode\": \"absolute\",\n            \"steps\": [\n              {\n                \"color\": \"green\",\n                \"value\": null\n              },\n              {\n                \"color\": \"red\",\n                \"value\": 1100\n              }\n            ]\n          },\n          \"unit\": \"short\"\n        },\n        \"overrides\": []\n      },\n      \"gridPos\": {\n        \"h\": 9,\n        \"w\": 12,\n        \"x\": 0,\n        \"y\": 8\n      },\n      \"id\": 2,\n      \"options\": {\n        \"colorMode\": \"background\",\n        \"graphMode\": \"none\",\n        \"justifyMode\": \"auto\",\n        \"orientation\": \"auto\",\n        \"reduceOptions\": {\n          \"calcs\": [\n            \"lastNotNull\"\n          ],\n          \"fields\": \"\",\n          \"values\": false\n        },\n        \"textMode\": \"auto\"\n      },\n      \"pluginVersion\": \"\",\n      \"targets\": [\n        {\n          \"datasource\": {\n            \"type\": \"prometheus\"\n          },\n          \"editorMode\": \"builder\",\n          \"expr\": \"sum by(region) (aws_quota_lambda_concurrent_executions)\",\n          \"legendFormat\": \"__auto\",\n          \"range\": true,\n          \"refId\": \"A\"\n        }\n      ],\n      \"title\": \"Lambda Execution concurrency\",\n      \"type\": \"stat\"\n    },\n    {\n      \"datasource\": {\n        \"type\": \"prometheus\"\n      },\n      \"fieldConfig\": {\n        \"defaults\": {\n          \"color\": {\n            \"mode\": \"continuous-GrYlRd\"\n          },\n          \"mappings\": [],\n          \"thresholds\": {\n            \"mode\": \"absolute\",\n            \"steps\": [\n              {\n                \"color\": \"green\",\n                \"value\": null\n              },\n              {\n                \"color\": \"red\",\n                \"value\": 80\n              }\n            ]\n          }\n        },\n        \"overrides\": []\n      },\n      \"gridPos\": {\n        \"h\": 8,\n        \"w\": 12,\n        \"x\": 12,\n        \"y\": 8\n      },\n      \"id\": 8,\n      \"options\": {\n        \"displayMode\": \"lcd\",\n        \"minVizHeight\": 10,\n        \"minVizWidth\": 0,\n        \"orientation\": \"horizontal\",\n        \"reduceOptions\": {\n          \"calcs\": [\n            \"lastNotNull\"\n          ],\n          \"fields\": \"\",\n          \"values\": false\n        },\n        \"showUnfilled\": true\n      },\n      \"pluginVersion\": \"\",\n      \"targets\": [\n        {\n          \"datasource\": {\n            \"type\": \"prometheus\"\n          },\n          \"editorMode\": \"builder\",\n          \"expr\": \"sum by(region) (aws_quota_lambda_function_and_layer_storage)\",\n          \"legendFormat\": \"__auto\",\n          \"range\": true,\n          \"refId\": \"A\"\n        }\n      ],\n      \"title\": \"Lambda Function & Layer Storage Limit\",\n      \"type\": \"bargauge\"\n    }\n  ],\n  \"schemaVersion\": 36,\n  \"style\": \"dark\",\n  \"tags\": [],\n  \"templating\": {\n    \"list\": []\n  },\n  \"time\": {\n    \"from\": \"now-6h\",\n    \"to\": \"now\"\n  },\n  \"timepicker\": {},\n  \"timezone\": \"\",\n  \"title\": \"Quotas\",\n  \"version\": 1,\n  \"weekStart\": \"\"\n}\n"` |  |
| grafana.datasources."datasources.yaml".apiVersion | int | `1` |  |
| grafana.datasources."datasources.yaml".datasources[0].access | string | `"proxy"` |  |
| grafana.datasources."datasources.yaml".datasources[0].editable | bool | `true` |  |
| grafana.datasources."datasources.yaml".datasources[0].name | string | `"prometheus"` |  |
| grafana.datasources."datasources.yaml".datasources[0].orgId | int | `1` |  |
| grafana.datasources."datasources.yaml".datasources[0].type | string | `"prometheus"` |  |
| grafana.datasources."datasources.yaml".datasources[0].url | string | `"http://prometheus.default.svc.cluster.local:9090"` |  |
| grafana.enabled | bool | `true` |  |
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
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
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
| replicaCount | int | `1` |  |
| resources | object | `{}` |  |
| secret.AWS_ACCESS_KEY_ID | string | `"QVdTX0FDQ0VTU19LRVlfSUQK"` |  |
| secret.AWS_SECRET_ACCESS_KEY | string | `"QVdTX1NFQ1JFVF9BQ0NFU1NfS0VZCg=="` |  |
| securityContext | object | `{}` |  |
| service.port | int | `10100` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `false` |  |
| serviceAccount.name | string | `""` |  |
| tolerations | list | `[]` |  |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.11.0](https://github.com/norwoodj/helm-docs/releases/v1.11.0)
