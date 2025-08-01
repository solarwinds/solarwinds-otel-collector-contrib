# SolarWinds Processor

| Status        |           |
| ------------- |-----------|
| Stability     | [development]: traces, metrics, logs   |
| Distributions | [verified], [playground], [k8s] |
| Issues        | [![Open issues](https://img.shields.io/github/issues-search/solarwinds/solarwinds-otel-collector-contrib?query=is%3Aissue%20is%3Aopen%20label%3Aprocessor%2Fsolarwinds%20&label=open&color=orange&logo=opentelemetry)](https://github.com/solarwinds/solarwinds-otel-collector-contrib/issues?q=is%3Aopen+is%3Aissue+label%3Aprocessor%2Fsolarwinds) [![Closed issues](https://img.shields.io/github/issues-search/solarwinds/solarwinds-otel-collector-contrib?query=is%3Aissue%20is%3Aclosed%20label%3Aprocessor%2Fsolarwinds%20&label=closed&color=blue&logo=opentelemetry)](https://github.com/solarwinds/solarwinds-otel-collector-contrib/issues?q=is%3Aclosed+is%3Aissue+label%3Aprocessor%2Fsolarwinds) |
| [Code Owners](https://github.com/solarwinds/solarwinds-otel-collector-contrib/blob/main/CODEOWNERS) |

[development]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#development
[verified]: https://github.com/solarwinds/solarwinds-otel-collector-releases/tree/main/distributions/verified
[playground]: https://github.com/solarwinds/solarwinds-otel-collector-releases/tree/main/distributions/playground
[k8s]: https://github.com/solarwinds/solarwinds-otel-collector-releases/tree/main/distributions/k8s

## General Description
SolarWinds processor provides modifying functionality to OTEL signals (metrics, logs, traces) and helps to gain full SolarWinds Observability experience.

## Feature Set
Currently there are two major functionalities provided by SolarWinds processor:
- Signals decoration by configured resource attributes. For this functionality `resource` configuration should be used. See some integration examples stored in [solarwinds-otel-collector-releases repository](https://github.com/solarwinds/solarwinds-otel-collector-releases/tree/main/examples/integrations).
- Checks on outgoing signal size in serialized form. Making sure our ingestion pipeline is capable of ingest signal. When serialized signal size is exceeded warning message is recorded in log. For this functionality `max_size_mib` property is used. Default value for this check is 6MiB. If there is a need to switch off signal size check, set `max_size_mib` to zero and check will be skipped.

## `solarwinds` Extension Requirement

`solarWinds` processor is tightly connected to existence of `solarWinds` extension. Processor uses part of configuration from `solarWinds` extension to properly decorate outgoing signals.

> [!WARNING]
> Exact name of extension is required in processor configuration. In case when `solarWinds` extension is misconfigured in processor setup, collector will not start.

## Default Resource Attributes
`solarwinds` processor implicitly decorates signals by following resource attributes. Their values are taken from `solarwinds` extension configuration.
- `sw.otelcol.receiver.name` - required. Controlled by `solarwinds` extension setup.
- `sw.otelcol.collector.entity_creation` - optional. Controlled by `solarwinds` extension setup.

## Processor Fully Explicit Configuration
Shows detailed `solarwinds` processor configuration. With special emphasis on binding to `solarwinds` extension.

Rest of configuration is heavily omitted to stay detailed (for the processor), but small (for all).

```yaml
extensions:
  solarwinds/sample:
    # Omitted for simplicity.
    
processors:
  solarwinds:
    # Needs to have exact name of configured extension.
    extension: solarwinds/sample
    # Checks on serialized signal size, in this case for 3MiB.
    max_size_mib: 3
    # Whatever resource attributes to be added.
    resource:
      attribute.name: "attribute_value"

service:
  extensions:
    - solarwinds/sample
  pipelines:
    metrics:
      processors:
        - solarwinds
```

## Minimal Processor Configuration
This configuration uses only mandatory configuration properties. Rest of properties are configured in default.

Following configuration provides no explicit resource attributes decoration. Signal size is checked against default 6MiB limit for serialized form.

```yaml
extensions:
  solarwinds/sample:
    # Omitted for simplicity.
    
processors:
  solarwinds:
    # Needs to have exact name of configured extension.
    extension: solarwinds/sample
```
