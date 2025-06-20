# Solarwinds Entity Connector

| Status        |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
|---------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Stability     | [development]: metrics_to_logs, logs_to_logs                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| Distributions | []                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| Issues        | [![Open issues](https://img.shields.io/github/issues-search/solarwinds/solarwinds-otel-collector-contrib?query=is%3Aissue%20is%3Aopen%20label%3Aconnector%2Fsolarwindsentity%20&label=open&color=orange&logo=opentelemetry)](https://github.com/solarwinds/solarwinds-otel-collector-contrib/issues?q=is%3Aopen%20is%3Aissue%20label%3Aconnector%2Fsolarwindsentity) [![Closed issues](https://img.shields.io/github/issues-search/solarwinds/solarwinds-otel-collector-contrib?query=is%3Aissue%20is%3Aclosed%20label%3Aconnector%2Fsolarwindsentity%20&label=closed&color=blue&logo=opentelemetry)](https://github.com/solarwinds/solarwinds-otel-collector-contrib/issues?q=is%3Aclosed%20is%3Aissue%20label%3Aconnector%2Fsolarwindsentity) |

[development]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#development

## Supported Data Type
| [Exporter Pipeline Type] | [Receiver Pipeline Type] |
|--------------------------|--------------------------|
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
        relationships:
          - context: metric
            conditions: 
              - metric.name == "k8s.tcp.bytes"
            type: CommunicatesWith
            source_entity: KubernetesPod
            destination_entity: KubernetesDeployment
            attributes:
              - sw.connection.status
```

### Configuration Options
#### Prefixes
- `source_prefix` and `destination_prefix` are used for all kinds of relationships.
  - For same-type relationships:
    - The connector expects source and destination resource attributes (IDs only) to be prefixed with `source.` and `dest.` respectively.
    - For example, if both entities have attributes `k8s.pod.name`, then the connector expects them to be prefixed as `source.k8s.pod.name` and `dest.k8s.pod.name` in resource attributes of incoming telemetry.
    - Not all attributes need to be prefixed, the requirement is to have at least one attribute prefixed with `source.` and one with `dest.` for the connector to identify the source and destination entities,
      to build valid event.
  - For different-type relationships:
    - Prefixes are supported for source and destination entity ID attributes, but are not required.
  - No defaults are provided for these prefixes, so they must be set explicitly in the configuration if same-type relationships are expected or the prefix is used for different-type relationships.

#### Schema
Defines the entities and relationships to be created/updated/deleted from incoming telemetry. To have action performed, the event
has to be defined in the `schema.events` section.
- `entities` is a list of entity definitions, with the following properties. All the property values have to be defined in SWO system for the specific entity.
  - `entity` type as defined in SWO,
  - `id` attributes are used as the identifiers, these have to match identification properties in SWO,
  - `attributes` are optional.

#### Events
Events define rules for creating entity/relationship from incoming telemetry. Both works with [conditions](#conditions) and `context` (log or metric).

- `entities` defines rules for creating entity events.
  - Entity is matched by the entity type.
  - ID attributes have to be present in the incoming telemetry as resource attributes.

- `relationships` defines relationships between entities.
  - The ID attributes of both entities must be present in the incoming telemetry as resource attributes. The expected
    ID attributes are found by looking into the source/destination entity definition.
  - expected attributes are:
    - relationship type,
    - source entity type,
    - destination entity,
    - relationship attributes (optional)


Entities referenced in the `events.entities` and `events.relationships` must be defined in the `schema.entities` section.


#### Conditions
Conditions are part of both `relationship` and `entity` events and use OTTL syntax. For each event 1+ conditions can be defined.

One condition can be composed (using operands like and, or, etc...) as OTTL format allows. In case there are multiple items in the conditions array, it will be evaluated as a logical **OR**:
  ```yaml
  conditions:
    - con1 or con2 and con3
    - con4
  ```

- If *no condition* items are specified, an event will always be created.
- `context` is required, `conditions` are optional
- You can find more information about OTTL paths and syntax examples in the [OTTL contexts documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/pkg/ottl/contexts)
