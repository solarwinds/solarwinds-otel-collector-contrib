# Recommendations for SolarWinds Entity Connector

## Logs Deduplication

To reduce the amount of incoming and outgoing logs and optimize telemetry load for connector,
you can use the log deduplication processor in collector configuration.
This processor helps eliminate duplicated log entries.


The configuration of logdedupprocessor provides various options to set up the log deduplication mechanism,
including the deduplication interval and conditions for filtering logs. 
The processor waits for the specified interval before processing logs to ensure that duplicate entries are filtered out.

For more details and options,
refer to the [logdedupprocessor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/logdedupprocessor) documentation.

### Example Configuration
For example, if `conditions` are defined, the log deduplication processor will only apply to logs that match the specified conditions.
Other logs will be processed normally without deduplication. Note that the `conditions` are joined with `OR` operator.
```yaml
processors:
  logdedup/example:
    interval: 20s
    conditions:
      - resource.attributes["k8s.namespace.name"] == "src-namespace"
```


### Reducing Incoming Telemetry
To further reduce the volume of incoming telemetry data, you can use the `logdedupprocessor` in your OpenTelemetry Collector configuration.
#### Pipeline Configuration Example
To apply the log deduplication processor to incoming logs, use the `logdedupprocessor` before the `solarwindsentity` receiver.
```yaml
    logs/out:
      receivers: [ ... ]
      processors: [ logdedup/input ]
      exporters: [ solarwindsentity ]
```

### Reducing Outgoing Telemetry
To reduce the volume of outgoing telemetry data, you can also apply the `logdedupprocessor` to the logs 
before they are sent to the SolarWinds Observability.

#### Pipeline Configuration Example
To apply the log deduplication processor to outgoing logs use logdedupprocessor after the `solarwindsentity` receiver.
```yaml
logs/out:
  receivers: [ solarwindsentity ]
  processors: [ logdedup/output ]
  exporters: [ ... ]
```

#### Usage Examples
- Deduplicate identical logs carrying updates for Kubernetes Namespace entity. Resulting resource logs will contain 
    log record per each entity received in 20s buckets.
    ```yaml
  logdedup/output:
    interval: 20s
    conditions:
      - attributes["otel.entity.type"] == "KubernetesNamespace"
    ```
- Deduplicate logs for a specific Kubernetes Namespace entity. Logs will be deduplicated for specified entity in 10s buckets.
    ```yaml
  logdedup/output:
      interval: 10s
      conditions:
        - attributes["otel.entity.id"]["sw.k8s.cluster.uid"] == "cluster-name" and attributes["otel.entity.id"]["k8s.namespace.name"] == "namespace"
    ```
  
See [logs format](OUTPUT.md) for more details on the log structure and its accessors.