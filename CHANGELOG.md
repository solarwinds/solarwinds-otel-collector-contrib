# Changelog

## v0.113.3
Removes `insecure` testing configuration parameter for [SolarWinds Extension](./extension/solarwindsextension).

## v0.113.2
Fixes OTLP port number used for exporting telemetry.

## v0.113.1
Adds [SolarWinds Extension](./extension/solarwindsextension). The [SolarWinds Exporter](./exporter/solarwindsexporter) is now dependent on the extension.

## v0.113.0
Initial version of SolarWinds OpenTelemetry Collector.
The collector provides all available components (receivers, processors, exporters, connectors, providers)
from [opentelemetry-collector](https://github.com/open-telemetry/opentelemetry-collector/tree/v0.113.0) (version `v0.113.0`) and [opentelemetry-collector-contrib](https://github.com/open-telemetry/opentelemetry-collector/tree/v0.113.0) (version `v0.113.0`).

### Additional details:
- `solarwindsexporter` has been added to easily integrate with **SolarWinds Observability SaaS**. Please read its [documentation](exporter/solarwindsexporter/README.md) to learn more.
