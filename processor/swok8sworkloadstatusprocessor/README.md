# SolarWinds Kubernetes Workload Status Processor

[verified]: https://github.com/solarwinds/solarwinds-otel-collector-releases#verified
[playground]: https://github.com/solarwinds/solarwinds-otel-collector-releases#playground
[k8s]: https://github.com/solarwinds/solarwinds-otel-collector-releases#k8s

| Status        |           |
| ------------- |-----------|
| Stability     | [development]: logs   |
| Distributions | [k8s], [playground], [verified] |
| Issues        | [![Open issues](https://img.shields.io/github/issues-search/solarwinds/solarwinds-otel-collector-contrib?query=is%3Aissue%20is%3Aopen%20label%3Aprocessor%2Fswok8sworkloadstatus%20&label=open&color=orange&logo=opentelemetry)](https://github.com/solarwinds/solarwinds-otel-collector-contrib/issues?q=is%3Aopen+is%3Aissue+label%3Aprocessor%2Fswok8sworkloadstatus) [![Closed issues](https://img.shields.io/github/issues-search/solarwinds/solarwinds-otel-collector-contrib?query=is%3Aissue%20is%3Aclosed%20label%3Aprocessor%2Fswok8sworkloadstatus%20&label=closed&color=blue&logo=opentelemetry)](https://github.com/solarwinds/solarwinds-otel-collector-contrib/issues?q=is%3Aclosed+is%3Aissue+label%3Aprocessor%2Fswok8sworkloadstatus) |

[development]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#development
[k8s]: https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-k8s
[playground]: 
[verified]: 

The SolarWinds Kubernetes Workload Status Processor derives and annotates Kubernetes workload status information on log records based on:
- Direct inspection of Deployment manifest JSON contained in the log `body`
- Aggregated phase of Pods belonging to a DaemonSet or StatefulSet

Depending on the workload kind the following attributes may be added when a status is computed:
- `sw.k8s.deployment.status`:
  - **AVAILABLE**: `Available=True` AND `Progressing=True`, 
  - **OUTDATED**: `Available=True` AND `Progressing!=True`
  - **UNAVAILABLE**: `Available!=True`
  - **OTHER**:  Fallback
- `sw.k8s.daemonset.status` and `sw.k8s.statefulset.status`:
  - **RUNNING**: (Running pods + Succeeded pods) == Total pods
  - **FAILED**: (Failed pods + Unknown pods) == Total pods
  - *Empty Status*: No pods
  - **PENDING**: Fallback

## Configuration

### Example

No custom configuration is needed. You can simply use:
```yaml
processors:
  swok8sworkloadstatus: {}
```