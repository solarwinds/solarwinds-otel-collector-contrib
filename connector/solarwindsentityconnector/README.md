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
        relationships:
          - type: CommunicatesWith
            source_entity: KubernetesPod
            destination_entity: KubernetesDeployment
            attributes:
              - sw.connection.status
```

### Configuration Options
- `source_prefix` and `destination_prefix` are used for same-type relationships when source and destination attributes has to be correctly set to create expected relationship.
  - The `solarwindsentity` connector expects source and destination resource attributes (IDs only) to be prefixed with `source.` and `dest.` respectively.
  - For example, if both entities have attributes `k8s.pod.name`, then the connector expects them to be prefixed as `source.k8s.pod.name` and `dest.k8s.pod.name` in resource attributes of incoming telemetry.
  - No defaults are provided for these prefixes, so they must be set explicitly in the configuration if same-type relationships are expected.
- `schema` defines the entities and relationships to be created/updated/deleted from incoming telemetry.
  - `entities` is a list of entity definitions, with the following properties. All the property values have to be defined in Entity Service Schemas (ESS) for the specific entity and be marked with `@telemetryMapping`.
    - `entity` type as defined in ESS,
    - `id` attributes are used as the identifiers, these have to match properties marked as `hashKeyOrder` in the ESS,
    - `attributes` are optional.
  - `events.relationships` defines relationships between entities, specifying the relationship type, source entity, destination entity, and optional attributes.
    - Entities used in the relationships must be defined in the `entities` section.