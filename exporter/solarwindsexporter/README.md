# Solarwinds Exporter
<!-- status autogenerated section -->
| Status        |           |
| ------------- |-----------|
| Stability     | [development]: traces, metrics, logs   |
| Deprecation of solarwindsexporter | [Date]: 2025-07-30   |
|                      | [Migration Note]: For full parity with `solarwinds` exporter use combination of `solarwinds` processor and `otlp` exporter. See [solarwinds processor readme](../../processor/solarwindsprocessor/README.md) for more information.   |
| Distributions | [] |
| Issues        | [![Open issues](https://img.shields.io/github/issues-search/solarwinds/solarwinds-otel-collector-contrib?query=is%3Aissue%20is%3Aopen%20label%3Aexporter%2Fsolarwinds%20&label=open&color=orange&logo=opentelemetry)](https://github.com/solarwinds/solarwinds-otel-collector-contrib/issues?q=is%3Aopen+is%3Aissue+label%3Aexporter%2Fsolarwinds) [![Closed issues](https://img.shields.io/github/issues-search/solarwinds/solarwinds-otel-collector-contrib?query=is%3Aissue%20is%3Aclosed%20label%3Aexporter%2Fsolarwinds%20&label=closed&color=blue&logo=opentelemetry)](https://github.com/solarwinds/solarwinds-otel-collector-contrib/issues?q=is%3Aclosed+is%3Aissue+label%3Aexporter%2Fsolarwinds) |
| Code coverage | [![codecov](https://codecov.io/github/solarwinds/solarwinds-otel-collector-contrib/graph/main/badge.svg?component=exporter_solarwinds)](https://app.codecov.io/gh/solarwinds/solarwinds-otel-collector-contrib/tree/main/?components%5B0%5D=exporter_solarwinds&displayType=list) |

[development]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#development
[Date]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#deprecation-information
[Migration Note]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#deprecation-information
<!-- end autogenerated section -->

SolarWinds Exporter is a convenience wrapper around [OTLP gRPC Exporter](https://github.com/open-telemetry/opentelemetry-collector/blob/main/exporter/otlpexporter/README.md) to transport your telemetry to **SolarWinds Observability SaaS** in an easy-to-configure manner.

## Getting Started

You just need to include the SolarWinds Exporter in your exporter definitions and no additional configuration is needed. It always needs to be used together with the [Solarwinds Extension](../../extension/solarwindsextension).

```yaml
exporters:
  solarwinds:
extensions:
  solarwinds:
    token: "TOKEN"
    data_center: "na-01"
```

## Full configuration

### Example with Defaults
```yaml
exporters:
  solarwinds:
    extension: "solarwinds"
    timeout: "10s"
    sending_queue:
      enabled: true
      num_consumers: 10
      queue_size: 1000
    retry_on_failure:
      enabled: true
      initial_interval: "5s"
      randomization_factor: 0.5
      multiplier: 1.5
      max_interval: "30s"
      max_elapsed_time: "300s"
extensions:
  solarwinds:
    token: "TOKEN"
    data_center: "na-01"
```
> [!TIP]  
> You can omit `extension` from the Solarwinds Exporter configuration above if there's only a single instance of the Solarwinds Extension.

- `extension` (optional) - This name identifies an instance of the [Solarwinds Extension](../../extension/solarwindsextension) to be used by this exporter to obtain its configuration. 
   If there is only a single instance of the extension, the configuration value is optional. The format mimics the identifier as it occurs in the collector configuration - 
   `type/name`, e.g `solarwinds` or `solarwinds/1` for multiple instances of the extension. You would use multiple instances for publishing your telemetry to 
   multiple **SolarWinds Observability SaaS** organizations.
- `timeout` (optional) - Timeout for each attempt to send data to the SaaS service. A timeout of zero disables the timeout. The **default** is `10s`.
- `retry_on_failure` (optional) - These options configure the retry behavior. Please refer to the [Exporter Helper documentation](https://github.com/open-telemetry/opentelemetry-collector/blob/main/exporter/exporterhelper/README.md).
- `sending_queue` (optional) - These are the options to set queuing in the exporter. A full descriptions can be similarly found in [Exporter Helper documentation](https://github.com/open-telemetry/opentelemetry-collector/blob/main/exporter/exporterhelper/README.md).

> [!NOTE]  
> The format of all durations above follow the [time.ParseDuration](https://pkg.go.dev/time#ParseDuration) format of "Duration" strings.

### Example with Multiple Solarwinds Extensions
```yaml
exporters:
  solarwinds:
    extension: "solarwinds/1"
extensions:
  solarwinds/1:
    token: YOUR-INGESTION-TOKEN1"
    data_center: "na-01"
  solarwinds/2:
    token: YOUR-INGESTION-TOKEN2"
    data_center: "na-02"
```
> [!WARNING]
> The `extension` configuration value cannot be omitted in the example above. 
> There are multiple instances of the Solarwinds Extension and you need to
> configure which instance to use to obtain configuration for the exporter.

## Development
- **Tests** can be executed with `make test`.
- After changes to `metadata.yaml` generated files need to be re-generated with `make generate`. The [mdatagen](http://go.opentelemetry.io/collector/cmd/mdatagen) tool has to be in the `PATH`.
