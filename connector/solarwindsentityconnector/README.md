# Solarwinds Entity Connector

| Status        |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Stability     | [development]: metrics_to_logs, logs_to_logs                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| Distributions | []                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| Issues        | [![Open issues](https://img.shields.io/github/issues-search/solarwinds/solarwinds-otel-collector-contrib?query=is%3Aissue%20is%3Aopen%20label%3Aconnector%2Fsolarwindsentity%20&label=open&color=orange&logo=opentelemetry)](https://github.com/solarwinds/solarwinds-otel-collector-contrib/issues?q=is%3Aopen%20is%3Aissue%20label%3Aconnector%2Fsolarwindsentity) [![Closed issues](https://img.shields.io/github/issues-search/solarwinds/solarwinds-otel-collector-contrib?query=is%3Aissue%20is%3Aclosed%20label%3Aconnector%2Fsolarwindsentity%20&label=closed&color=blue&logo=opentelemetry)](https://github.com/solarwinds/solarwinds-otel-collector-contrib/issues?q=is%3Aclosed%20is%3Aissue%20label%3Aconnector%2Fsolarwindsentity) |

[development]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#development

## Supported Data Type

| [Exporter Pipeline Type] | [Receiver Pipeline Type] |
| ------------------------ | ------------------------ |
| metrics                  | logs                     |
| logs                     | logs                     |

Transforms metrics and logs to logs representing entity state updates/deletes and relationship updates/deletes. The entity connector should be used with SolarWinds exporter to send logs to SolarWinds Observability SaaS.

## Configuration

See example configuration below.

```yaml
connectors:
  solarwindsentity:
    source_prefix: "source."
    destination_prefix: "dest."
    expiration_policy:
      enabled: true
      interval: 5m
    schema:
      entities:
        - entity: KubernetesPod
          id:
            - k8s.pod.name
            - k8s.namespace.name
            - sw.k8s.cluster.uid
          attributes:
            - sw.k8s.pod.status
        - entity: KubernetesDeployment
          id:
            - k8s.deployment.name
            - k8s.namespace.name
            - sw.k8s.cluster.uid
      events:
        entities:
          - context: log
            type: KubernetesPod
          - context: metric
            type: KubernetesPod
            conditions:
              - metric.name == "k8s.tcp.bytes"
            action: "update"
        relationships:
          - context: metric
            conditions:
              - metric.name == "k8s.tcp.bytes"
            type: CommunicatesWith
            source_entity: KubernetesPod
            destination_entity: KubernetesDeployment
            attributes:
              - sw.connection.status
            action: "delete"
```

### Prefixes

- `source_prefix`
  - Required: no (default: "")
  - Description: Prefix for source entity ID attributes while detecting relationship events. See details bellow.
- `destination_prefix`
  - Required: no (default: "")
  - Description: Prefix for destination entity ID attributes while detecting relationship events. See details bellow.

`source_prefix` and `destination_prefix` are used for all kinds of relationships.

- For same-type relationships:
  - The connector expects source and destination resource attributes (IDs only) to be prefixed with `source.` and `dest.` respectively.
  - For example, if both entities have attributes `k8s.pod.name`, then the connector expects them to be prefixed as `source.k8s.pod.name` and `dest.k8s.pod.name` in resource attributes of incoming telemetry.
  - Not all attributes need to be prefixed, the requirement is to have at least one attribute prefixed with `source.` and one with `dest.` for the connector to identify the source and destination entities,
    to build valid event.
- For different-type relationships:
  - Prefixes are supported for source and destination entity ID attributes, but are not required.
- No defaults are provided for these prefixes, so they must be set explicitly in the configuration if same-type relationships are expected or the prefix is used for different-type relationships.

### Expiration Policy

- `expiration_policy` section defines the expiration policy for relationships.
  - `enabled`
    - Required: no (default: `false`)
    - Description: Enables the expiration policy.
  - `interval`
    - Required: no (default: `5m`)
    - Description: Defines TTL of the relationships in time.Duration format (e.g., `5m` for 5 minutes). After this interval, the relationship is considered expired and a delete event is sent to SolarWinds Observability SaaS.
  - `cache_configuration` section defines the cache configuration.
    - `ttl_cleanup_interval`
      - Required: no (default: `5s`)
      - Description: Defines how often the cache is cleaned up in time.Duration format (e.g., `5m` for 5 minutes). This is the time when the expired relationships are removed from the cache and delete events are sent to SolarWinds Observability SaaS. Minimum value is `1s`.
    - `max_capacity`
      - Required: no (default: `1 000 000`)
      - Description: Defines the maximum number of relationships that can be stored in the cache.

### Schema

- `schema` section defines the entities and relationships to be created/updated/deleted from incoming telemetry. To have action performed, the event has to be defined in the [`schema.events`](#events) section.

#### Entities

- `schema.entities` is a list of entity definitions. All the property values have to be defined in SWO system for the specific entity.
  - `entity`
    - Required: yes
    - Description: Entity type as defined in SWO.
  - `id`
    - Required: yes
    - Description: List of attributes used as the identifiers. These have to match identification properties in SWO.
  - `attributes`
    - Required: no (default: empty)
    - Description: List of entity attributes.

#### Events

- `schema.events` define rules for creating entity/relationship from incoming telemetry. Both work with [conditions](#conditions) and `context` (`log` or `metric`). Events have two possible `action` types: `update` and `delete`. Use the corresponding action to `update` (update or insert if new) or `delete` an entity/relationship in SolarWinds Observability SaaS.
  - `entities` defines rules for creating entity events. Entity is matched by the entity type. ID attributes have to be present in the incoming telemetry as resource attributes.
    - `action`
      - Required: yes
      - Description: Action type for the event. Must be either `update` or `delete`.
    - `context`
      - Required: yes
      - Description: Context type for the event. Must be either `log` or `metric`.
    - `entity`
      - Required: yes
      - Description: Entity name that must be defined in the `schema.entities` section.
    - `conditions`
      - Required: no (default: `true`)
      - Description: List of OTTL conditions. See [details](#conditions) bellow.
  - `relationships` defines relationships between entities. The ID attributes of both entities must be present in the incoming telemetry as resource attributes. The expected ID attributes are found by looking into the source/destination entity definition.
    - `action`
      - Required: yes
      - Description: Action type for the relationship event. Must be either `update` or `delete`.
    - `context`
      - Required: yes
      - Description: Context type for the relationship event. Must be either `log` or `metric`.
    - `type`
      - Required: yes
      - Description: Relationship type as defined in SWO.
    - `source_entity`
      - Required: yes
      - Description: Source entity name that must be defined in the `schema.entities` section.
    - `destination_entity`
      - Required: yes
      - Description: Destination entity name that must be defined in the `schema.entities` section.
    - `conditions`
      - Required: no (default: `true`)
      - Description: List of OTTL conditions. See [details](#conditions) bellow.
    - `attributes`
      - Required: no (default: empty)
      - Description: List of relationship attributes.

**Note:** Entities referenced in the `events.entities` and `events.relationships` must be defined in the `schema.entities` section.

### Conditions

Conditions are part of both `relationship` and `entity` events and use OTTL syntax. For each event 1+ conditions can be defined.

One condition can be composed (using operands like and, or, etc...) as OTTL format allows. In case there are multiple items in the conditions array, it will be evaluated as a logical **OR**:

```yaml
conditions:
  - con1 or con2 and con3
  - con4
```

**Additional Information:**

- If no condition items are specified, an event will always be created.
- You can find more information about OTTL paths and syntax examples in the [OTTL contexts documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/pkg/ottl/contexts).
- Standard converters are supported, as described here: [Converters](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/pkg/ottl/ottlfuncs#converters). Note that the newest converters listed there might not be consumed by us yet.

## Recommendations

See [RECOMMENDATIONS](docs/RECOMMENDATIONS.md) for more information on how to use the SolarWinds Entity Connector effectively.

## Generated Logs

See [OUTPUT](docs/OUTPUT.md) for more information on the output format of the SolarWinds Entity Connector.
