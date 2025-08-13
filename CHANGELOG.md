# Changelog

## vNext

## v0.131.2
- Updates golang to 1.24.6
- Added `processesscraper` to `swohostmetricsreceiver` providing `swo.system.processes.count` metric

## v0.131.1
No changes, issues with previous release.

## v0.131.0
- Updated OpenTelemetry modules to [v1.37.0/v0.131.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.131.0)
- Updated TLS configuration in extensions and exporters (`TLSSetting` → `TLS`)
- Wrapped Keepalive configurations with `configoptional.Some()`
- Replaced deprecated `exporterhelper.QueueConfig` with `exporterhelper.QueueBatchConfig`
- Components affected: `solarwindsextension`, `solarwindsexporter`

## v0.127.9
- Chores without impact (excluded internal tools from CodeQL analysis, added dependency review action with GHAS)
- Fix CVE-2025-54388: Moby firewall reload could expose localhost-only containers to external access

## v0.127.8
- Updates golang to 1.24.5
- Chores without impact (version bumps, refactors and build pipeline improvements)

## v0.127.7
No changes, previous release failed to finish properly.

## v0.127.6
- `solarwindsentityconnector` Added support for prefixes on entities when entities are created from relationships
- `solarwindsentityconnector` Added support for OTTL converters in event condition expressions
- `solarwindsentityconnector` Fixed extra output log issue by adding missing yaml tags for cache configuration parsing
- ⚠️ Breaking change: `solarwindsentityconnector` Added configuration validation - a lot of properties are now required.
- ⚠️ Breaking change: `solarwindsentityconnector` The entities used in the `schema.events` (both, in entity and relationship events) must be defined in the `schema.entities` section.
- ⚠️ Breaking change: `solarwindsentityconnector` - `schema.events.entities[].type` renamed to `schema.events.entities[].entity`
- `swok8sworkloadtypeprocessor` Support addresses ending with dot
- `swohostmetricsreceiver` Separated `wmi` and `registry` packages in their own modules
- `k8seventgenerationprocessor` Extend the k8seventgeneration processor to extract data for Service mapping

## v0.127.5
- `solarwindsentityconnector` Benchmark fix
- Fix GHSA-fv92-fjc5-jj9h: mapstructure May Leak Sensitive Information in Logs When Processing Malformed Data

## v0.127.4
- `solarwindsprocessor` Introduced as replacement for `solarwinds` exporter.
- `solarwindsextension` Configuration extended by gRPC setup re-usable in `otlp` exporter. Some of current properties (`data_center`, `token` and `endpoint_url_override`) have been made deprecated and will be removed by July 30.
- `solarwindsexporter` Made deprecated and will be removed by July 30.
- `solarwindsentityconnector` Added benchmark tests to check the connector’s performance.
- `solarwindsentityconnector` Added support for delete events. Entities and relationships can be now deleted based on OTTL conditions.
- `solarwindsentityconnector` Relationships now expire (configurable), and a delete event is sent after the expiration.
- `swoworkloadtypeprocessor` Extended the processor to allow searching for owners of Pods.

## v0.127.3
- `swok8sobjectsreceiver` Change logging of 410 events to debug level.

## v0.127.2
- `solarwindsentityconnector` Added optional prefixes support for entity relationships between entities of different types.

## v0.127.1
- Various improvements to `swoworkloadtypeprocessor`:
  - Prevents overwriting workload type attribute if it's already set
  - Supports multiple contexts for attributes (resource, datapoint, metric, scope)
  - Supports searching for workload type by DNS and IP address
- `solarwindsentityconnector` supports OTTL conditions for conditional creation/update of entities and relationships 
- Utilizes reusable build components from [solarwinds-otel-collector-core](https://github.com/solarwinds/solarwinds-otel-collector-core)

## v0.127.0
- Consumes OpenTelemetry Collector dependencies v0.127.0.

## v0.123.7
- Aligning version with [solarwinds-otel-collector-releases](https://github.com/solarwinds/solarwinds-otel-collector-releases) repository.
- Latest version of `solarwindsentityconnector` moved to this repository from [solarwinds-otel-collector-releases](https://github.com/solarwinds/solarwinds-otel-collector-releases).

## v0.123.4
- Release of ./internal/k8sconfig.

## v0.123.3
- Fixing module name for SolarWinds entity connector.

## v0.123.2
- Moved connection-check code to separate binary. Binary is added to k8s docker images.
- Adds [SolarWinds Kubernetes Workload Type Processor](./processor/swok8sworkloadtypeprocessor/README.md) for annotating metrics with a k8s workload type based on their attributes.

## v0.123.1
- Fix CVE-2025-22872: golang.org/x/net vulnerable to Cross-site Scripting

## v0.123.0
- Consumes OpenTelemetry Collector dependencies v0.123.0.
- SolarWinds exporter is now reported as otlp/solarwinds-<name> in collector's telemetry.

## v0.119.12
- Updates non-opentelemetry dependencies to latest possible version
- Sets metrics scope name to `github.com/solarwinds/solarwinds-otel-collector-releases`

## v0.119.11
- Fix CVE-2025-27144: Uncontrolled Resource Consumption

## v0.119.10
- Ignores any timestamps in all the Kubernetes manifests
- Fix CVE-2025-30204

## v0.119.9
- Fix CVE-2025-29786

## v0.119.8
- Adds ap-01 datacell support

## v0.119.7
- Fix CVE-2025-22866

## v0.119.6
- Updates Go build toolchain to 1.23.6

## v0.119.5
- Updating `swok8sobjectsreceiver` to remove `managedFields` for both PULL and WATCH objects

## v0.119.4
- Updating `swok8sobjectsreceiver` to report changes in other than `status`, `spec`, and `metadata` sections

## v0.119.3
- Adds custom `k8sobjectsreceiver` to notify about what sections in manifest were changed 

## v0.119.2
- Fix CVE-2025-22869
- Fix CVE-2025-22868

## v0.119.1
- Utilizes `pdatatest` for specific E2E tests.
- SolarWinds-specific packages are tagged and can be referenced from other repositories.
- Adds custom `k8seventgenerationprocessor` to transform K8S entities change events to logs.
- Removes opentelemetry-collector wrapper used for corruption prevention as newly introduced `fsync@v0.96.0` solves the issue.

## v0.119.0
- Consumes OpenTelemetry Collector dependencies v0.119.0.

## v0.113.8
- Updates Go build toolchain to 1.23.5.
- Adds [SWO Host Metrics Receiver](./receiver/swohostmetricsreceiver/README.md) for additional Host metrics monitoring.
- Adds connection check functionality to K8s distribution startup.
- Adds Windows architecture for Docker builds.

## v0.113.7
- Adds `without_entity` to [SolarWinds Extension](./extension/solarwindsextension/README.md#getting-started) configuration, so users can opt out of collector entity creation.
- Tags all signals with `entity_creation` attribute, except when without_entity is set on [SolarWinds Extension](./extension/solarwindsextension/README.md#getting-started).

## v0.113.6
- Marks all outgoing telemetry from the [SolarWinds Exporter](./exporter/solarwindsexporter) with
an attribute storing the collector name (`sw.otelcol.collector.name`) as it is configured in the
[SolarWinds Extension](./extension/solarwindsextension/README.md#getting-started).
- The uptime metric used to signal heartbeat is now decorated with `sw.otelcol.collector.version` which contains collector version.

## v0.113.5
Tags released docker images with `latest` tag.

## v0.113.4
Adds optional `resource` configuration parameter for [SolarWinds Extension](./extension/solarwindsextension).

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
