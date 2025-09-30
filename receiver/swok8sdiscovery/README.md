## swok8sdiscovery Receiver

| Status        |           |
| ------------- |-----------|
| Stability     | [alpha]: logs   |
| Distributions | [k8s] |
| Issues        | [![Open issues](https://img.shields.io/github/issues-search/solarwinds/solarwinds-otel-collector-contrib?query=is%3Aissue%20is%3Aopen%20label%3Areceiver%2Fswok8sdiscovery%20&label=open&color=orange&logo=opentelemetry)](https://github.com/solarwinds/solarwinds-otel-collector-contrib/issues?q=is%3Aopen+is%3Aissue+label%3Areceiver%2Fswok8sdiscovery) [![Closed issues](https://img.shields.io/github/issues-search/solarwinds/solarwinds-otel-collector-contrib?query=is%3Aissue%20is%3Aclosed%20label%3Areceiver%2Fswok8sdiscovery%20&label=closed&color=blue&logo=opentelemetry)](https://github.com/solarwinds/solarwinds-otel-collector-contrib/issues?q=is%3Aclosed+is%3Aissue+label%3Areceiver%2Fswok8sdiscovery) |

[alpha]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#alpha
[k8s]: https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-k8s



The `swok8sdiscovery` receiver performs periodic discovery of databases used in a Kubernetes cluster and emits **entity events as OpenTelemetry log records** describing discovered database instances and their relationships to owning Kubernetes workloads (Deployment / StatefulSet / DaemonSet / Job / CronJob).

Discovery currently supports two complementary strategies (both optional – enable either or both):

1. Image-based discovery (`database.image_rules`):
	 - Matches container image names against user-provided regular expressions.
	 - Optionally constrains to a single default port when specified.
	 - Resolves a stable endpoint using the best matching Service (selector overlaps chosen ports) or falls back to the Pod name.
2. Domain-based discovery (`database.domain_rules`):
	 - Matches `ExternalName` Services whose external DNS name matches configured patterns.
	 - When multiple rules match, the one whose `database_type` or any of its `domain_hints` appears in either the service name or external domain is preferred.

Each discovered database produces:
* An entity state log (type = `entity_state`) with attributes under `otel.entity.id` identifying the database (`sw.discovery.dbo.address`, `sw.discovery.dbo.type`, `sw.discovery.id`).
* (If workload ownership resolved) A relationship log (type = `entity_relationship_state`) linking the database entity to a Kubernetes workload (relation type `DiscoveredBy`).

### Emitted Attributes (selection)
| Attribute | Description |
|-----------|-------------|
| `otel.entity.event.type` | `entity_state` or `entity_relationship_state` |
| `otel.entity.type` | Always `DiscoveredDatabaseInstance` for entity events |
| `sw.discovery.dbo.address` | Database endpoint |
| `sw.discovery.dbo.port` | Exact port on which database is running |
| `sw.discovery.dbo.possible.ports` | Detected ports on database endpoint |
| `sw.discovery.dbo.type` | Logical database type (e.g. `mongo`, `postgres`, `redis`) |
| `sw.discovery.dbo.name` | Endpoint plus workload name (`<endpoint>#<workload>`) when workload present |
| `sw.discovery.source` | Value of configured `reporter` (for provenance) |
| `k8s.<workload kind>.name` | Name of owning workload (when resolved) |
| `k8s.namespace.name` | Namespace of the workload/pod/service |
| `sw.k8s.cluster.uid` | Cluster UID (from environment `CLUSTER_UID`) |

### Configuration

Top-level settings:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `interval` | duration | `5m` | Time between discovery cycles. Shorten in tests (e.g. `15s`). |
| `reporter` | string | empty | Optional source label recorded as `sw.discovery.source`. |
| `k8s` auth fields | (inlined via `APIConfig`) | | Standard Kubernetes client auth (service account, kubeconfig, etc.). |
| `database` | object | nil | Enables database discovery if provided. |

`database.image_rules` entries:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `database_type` | string | yes | Logical database type label. |
| `patterns` | []string (regex) | yes | Regex patterns matched against full container image (e.g. `docker.io/library/mongo:.*`). |
| `default_port` | int | no | If present and exists among container ports, only that port will be emitted (deduping multi-port images). |

`database.domain_rules` entries:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `database_type` | string | yes | Logical database type label. |
| `patterns` | []string (regex) | yes | Patterns matched against `ExternalName` value. |
| `domain_hints` | []string | no | Tie-break hints (substring matches in service name or external domain). |

### Example Configuration

```yaml
receivers:
	swok8sdiscovery:
		interval: 30s
		reporter: "agent"
		# Kubernetes auth (service account in-cluster example)
		auth_type: serviceAccount
		database:
			image_rules:
				- database_type: mongo
					patterns: [".*/mongo:.*"]
					default_port: 27017
				- database_type: postgres
					patterns: [".*/postgres:.*"]
					default_port: 5432
			domain_rules:
				- database_type: redis
					patterns: [".*redis.example.com"]

exporters:
	debug:
		verbosity: detailed

service:
	pipelines:
		logs:
			receivers: [swok8sdiscovery]
			exporters: [debug]
```
