# SolarWinds Entity Connector Behavior Examples
This document provides examples of how to configure the SolarWinds Entity Connector correctly to prevent inferring unexpected entities or relationships.

Each example includes:
- **Defined events**: Configuration specifying which entity types/relationships should be inferred and under what conditions
- **Input resource attributes**: Attributes present in the incoming telemetry resource  
- **Expected output**: Log records that will be sent to SolarWinds Observability

## Table of Contents
- [Entity ID Attributes](#entity-id-attributes)
  - [Single Entity](#single-entity)
    - [Without Prefix](#without-prefix)
    - [With Prefix](#with-prefix)
  - [Multiple Entities](#multiple-entities)
    - [Without Prefix](#without-prefix-1)
    - [Entities With the Same Set of IDs](#entities-with-the-same-set-of-ids)
    - [Entities With the Same Set of IDs Using Conditions](#entities-with-the-same-set-of-ids-using-conditions)
- [Relationship ID Attributes](#relationship-id-attributes)
  - [Different-Type Relationship](#different-type-relationship)
    - [Without Prefix](#without-prefix-2)
    - [With Prefix](#with-prefix-1)
  - [Same-Type Relationship](#same-type-relationship)
    - [With Partial Prefix](#with-partial-prefix)
    - [With Prefix](#with-prefix-2)
  - [Any-Type Relationship](#any-type-relationship)
    - [Without Inferring of Entities](#without-inferring-of-entities)
    - [Multiple Relationships](#multiple-relationships)
- [Entity Attributes](#entity-attributes)
  - [Without Prefix](#without-prefix-3)
  - [With Prefix](#with-prefix-3)
- [Relationship Attributes](#relationship-attributes)
  - [Without Prefix](#without-prefix-4)
  - [With Prefix and Entities](#with-prefix-and-entities)

## Entity Definitions Used in Examples
Examples below rely on the following entity definitions and schema configuration, including source/destination prefixes.

```yaml
source_prefix: "src."
destination_prefix: "dst."
schema:
  entities:
    - type: KubernetesCluster
      id: [cluster.uid]
      attributes: [cluster.name]
    - type: KubernetesNamespace
      id: [cluster.uid, namespace.name]
    - type: Snowflake
      id: [receiver.id, receiver.name]
    - type: DockerDaemon
      id: [receiver.id, receiver.name]
```

## Notes
- Resource Attributes are of `map` type so no duplicate entry is expected.
- `EXAMPLES.md` describes scenarios with simplified syntax, for supported syntax see [README.md](../README.md)
or [OUTPUT.md](OUTPUT.md) to see the actual output of the SolarWinds Entity Connector.
- `conditions: ["true"]` is the same as not defining conditions at all.

# Examples
## Entity ID Attributes
### Single Entity
#### Without Prefix
Entity will be inferred, because conditions are true and all ID attributes are found in the resource attributes.

**Defined events**

```yaml
events:
  entities:
    - type: KubernetesCluster
      action: update
      context: log
      conditions: ["resource.attributes['cluster.uid'] == 'cluster-123'"]
```

**Input resource attributes**

```
cluster.uid -> "cluster-123"
```


**Output log records**
```json
[
  {
    "entity_type": "KubernetesCluster",
    "entity_id": {
      "cluster.uid": "cluster-123"
    }
  }
]
```

___


#### With Prefix
Entity will NOT be inferred, because the prefix is processed for relationships and entities participating as source 
or destination. The entity would be inferred if valid relationship is inferred.

**Defined events**

```yaml
events:
  entities:
    - type: KubernetesCluster
      action: update
      context: log
      conditions: ["true"]
```

**Input resource attributes**

```
src.cluster.uid -> "cluster-123"
```

**Output log records**
```json
[]
```
___
### Multiple Entities
#### Without Prefix
Two entities will be inferred, because conditions are true and all ID attributes for both entities are found in 
the resource attributes. To send entity log updates for any entity, it has to be defined in the `events.entities` 
section.


**Defined events**

```yaml
events:
  entities:
    - type: KubernetesCluster
      action: update
      context: log
      conditions: ["true"]
    - type: KubernetesNamespace
      action: update
      context: log
      conditions: ["true"]
```

**Input resource attributes**

```
cluster.uid -> "cluster-123"
namespace.name -> "namespace-123"
```

**Output log records**
```json
[
  {
    "entity_type": "KubernetesCluster",
    "entity_id": {
      "cluster.uid": "cluster-123"
    }
  },
  {
    "entity_type": "KuberenetesNamespace",
    "entity_id": {
      "cluster.uid": "cluster-123",
        "namespace.name": "namespace-123"
    }
  }
]
```
___

#### Entities With the Same Set of IDs
:warning: Two entities will be inferred, because conditions are true and all ID attributes for both entities are found, even if
it is unexpected. To solve this issue, use `conditions` to filter out entities that should not be inferred (see scenario below).

**Defined events**

```yaml
events:
  entities:
    - type: Snowflake
      action: update
      context: log
      conditions: ["true"]
    - type: DockerDaemon
      action: update
      context: log
      conditions: ["true"]
```

**Input resource attributes**

```
receiver.id -> "snowflake-123"
receiver.name -> "snowflake"
```

**Output log records**
```json
[
  {
    "entity_type": "Snowflake",
    "entity_id": {
      "receiver.id": "snowflake-123",
      "receiver.name": "snowflake"
    }
  },
  {
    "entity_type": "DockerDaemon",
    "entity_id": {
      "receiver.id": "snowflake-123",
      "receiver.name": "snowflake"
    }
  }
]
```
___

#### Entities With the Same Set of IDs Using Conditions
One entity will be inferred, because conditions are true only for Snowflake entity.

**Defined events**

```yaml
events:
  entities:
    - type: Snowflake
      action: update
      context: log
      conditions: ["receiver.name == 'snowflake'"]
    - type: DockerDaemon
      action: update
      context: log
      conditions: ["receiver.name == 'docker'"]
```

**Input resource attributes**

```
receiver.id -> "snowflake-123"
receiver.name -> "snowflake"
```

**Output log records**
```json
[
  {
    "entity_type": "Snowflake",
    "entity_id": {
      "receiver.id": "snowflake-123",
      "receiver.name": "snowflake"
    }
  }
]
```
___
## Relationship ID Attributes
### Different-Type Relationship
#### Without Prefix
Relationship without prefixed attributes can be used to infer relationships between entities of different types (where entity IDs are not the same for each entity).

This configuration will output three log records. One for relationship, two for entities, because they were configured in events section.


**Defined events**

```yaml
events:
  entities:
    - type: KubernetesCluster
      action: update
      context: log
    - type: KubernetesNamespace
      action: update
      context: log
  relationships:
    - type: Has
      action: update
      context: log
      source_entity: KubernetesCluster
      destination_entity: KubernetesNamespace
```

**Input resource attributes**

```
cluster.uid -> "cluster-123"
namespace.name -> "namespace-123"
```

**Output log records**
```json
[
  {
    "entity_type": "KubernetesCluster",
    "entity_id": {
      "cluster.uid": "cluster-123"
    }
  },
  {
    "entity_type": "KubernetesNamespace",
    "entity_id": {
      "cluster.uid": "cluster-123",
        "namespace.name": "namespace-123"
    }
  },
  {
    "relationship_type": "Has",
    "source_entity_type": "KubernetesCluster",
    "source_entity_id": {
      "cluster.uid": "cluster-123"
    },
    "destination_entity_type": "KubernetesNamespace",
    "destination_entity_id": {
      "cluster.uid": "cluster-123",
      "namespace.name": "namespace-123"
    }
  }
]
```
___

#### With Prefix
Relationship with prefixed attributes can be used to infer relationships between entities of different types with or without same set of IDs, (this example shows with same set of IDs = Snowflake -> DockerDaemon).

Together with inferring entities, from prefixed attributes. However, in this scenario, we present case where
conditions for DockerDaemon entity are not met, so only Snowflake is inferred (no conditions means *true*).

This configuration will output two log records. One for relationship, one for Snowflake entity.


**Defined events**

```yaml
events:
  entities:
    - type: Snowflake
      action: update
      context: log
    - type: DockerDaemon
      action: update
      context: log
      conditions: ["something falsy"]
  relationships:
    - type: Has
      action: update
      context: log
      source_entity: Snowflake
      destination_entity: DockerDaemon
```

**Input resource attributes**

```
src.receiver.id -> "snowflake-123"
src.receiver.name -> "snowflake"
dst.receiver.id -> "docker-123"
dst.receiver.name -> "docker"
```

**Output log records**
```json
[
  {
    "entity_type": "Snowflake",
    "entity_id": {
      "receiver.id": "snowflake-123"
    }
  },
  {
    "relationship_type": "Has",
    "source_entity_type": "Snowflake",
    "source_entity_id": {
      "receiver.id": "snowflake-123",
      "receiver.name": "snowflake"
    },
    "destination_entity_type": "DockerDaemon",
    "destination_entity_id": {
      "receiver.id": "docker-123",
      "receiver.name": "docker"
    }
  }
]
```
___

### Same-Type Relationship
Same-type relationship needs at least one prefixed attribute to differ between source and destination entity.
#### With Partial Prefix
Relationship with attributes where at least one prefix is supported. All other attributes will be taken from
the unprefixed attributes and will behave like common attributes for both entities.


**Defined events**

```yaml
events:
  entities:
    - type: KubernetesNamespace
      action: update
      context: log
  relationships:
    - type: CommunicatesWith
      action: update
      context: log
      source_entity: KubernetesNamespace
      destination_entity: KubernetesNamespace
```

**Input resource attributes**

```
cluster.uid -> "cluster-123"
src.namespace.name -> "source namespace"
namespace.name -> "destination namespace"
```

**Output log records**
```json
[
  {
    "entity_type": "KubernetesNamespace",
    "entity_id": {
      "cluster.uid": "cluster-123",
      "namespace.name": "source namespace"
    }
  },
  {
    "entity_type": "KubernetesNamespace",
    "entity_id": {
      "cluster.uid": "cluster-123",
      "namespace.name": "destination namespace"
    }
  },
  {
    "relationship_type": "CommunicatesWith",
    "source_entity_type": "KubernetesNamespace",
    "source_entity_id": {
      "cluster.uid": "cluster-123",
      "namespace.name": "source namespace"
    },
    "destination_entity_type": "KubernetesNamespace",
    "destination_entity_id": {
      "cluster.uid": "cluster-123",
      "namespace.name": "destination namespace"
    }
  }
]
```
___


#### With Prefix
Relationship with prefixed attributes can be used to infer relationships between entities of the same type.

Together with inferring entities, from prefixed attributes.

This configuration will output three log records. One for relationship, two for entities of KubernetesCluster type,
because the type is configured in the events section.


**Defined events**

```yaml
events:
  entities:
    - type: KubernetesCluster
      action: update
      context: log
  relationships:
    - type: CommunicatesWith
      action: update
      context: log
      source_entity: KubernetesCluster
      destination_entity: KubernetesCluster
```

**Input resource attributes**

```
src.cluster.uid -> "cluster-123"
dst.cluster.uid -> "cluster-456"
```

**Output log records**
```json
[
  {
    "entity_type": "KubernetesCluster",
    "entity_id": {
      "cluster.uid": "cluster-123"
    }
  },
  {
    "entity_type": "KubernetesCluster",
    "entity_id": {
      "cluster.uid": "cluster-456"
    }
  },
  {
    "relationship_type": "CommunicatesWith",
    "source_entity_type": "KubernetesCluster",
    "source_entity_id": {
      "cluster.uid": "cluster-123"
    },
    "destination_entity_type": "KubernetesCluster",
    "destination_entity_id": {
      "cluster.uid": "cluster-456"
    }
  }
]
```
___



### Any-Type Relationship
Scenarios in this section are applicable to both types of relationships (same/different).
#### Without Inferring of Entities
This scenario presents case where only relationship is inferred without entities. Entities would be inferred
if mentioned in `events.entities`.

**Defined events**

```yaml
events:
  relationships:
    - type: Has
      action: update
      context: log
      source_entity: KubernetesCluster
      destination_entity: KubernetesNamespace
```

**Input resource attributes**

```
cluster.uid -> "cluster-123"
namespace.name -> "namespace-123"
```

**Output log records**
```json
[
  {
    "relationship_type": "Has",
    "source_entity_type": "KubernetesCluster",
    "source_entity_id": {
      "cluster.uid": "cluster-123"
    },
    "destination_entity_type": "KubernetesNamespace",
    "destination_entity_id": {
      "cluster.uid": "cluster-123",
      "namespace.name": "namespace-123"
    }
  }
]
```
___
#### Multiple Relationships
If conditions are fulfilled and attributes can be mapped to entities participating in relationships, more than one relationship can be inferred.

This scenario has one limitation: the entities have to have different sets of IDs, or at least be correctly differentiated by prefix. However, this should not be a usual case when taking into account the real usage of OpenTelemetry and the appearance of the incoming telemetry resources.

**Defined events**

```yaml
events:
  relationships:
    - type: Has
      action: update
      context: log
      source_entity: KubernetesCluster
      destination_entity: KubernetesNamespace
    - type: CommunicatesWith
      action: update
      context: log
      source_entity: Snowflake
      destination_entity: Snowflake
```

**Input resource attributes**

```
cluster.uid -> "cluster-123"
namespace.name -> "namespace-123"
src.receiver.id -> "snowflake-123"
dst.receiver.id -> "snowflake-456"
receiver.name -> "snowflake"
```

**Output log records**
```json
[
  {
    "relationship_type": "Has",
    "source_entity_type": "KubernetesCluster",
    "source_entity_id": {
      "cluster.uid": "cluster-123"
    },
    "destination_entity_type": "KubernetesNamespace",
    "destination_entity_id": {
      "cluster.uid": "cluster-123",
      "namespace.name": "namespace-123"
    }
  },
  {
    "relationship_type": "CommunicatesWith",
    "source_entity_type": "Snowflake",
    "source_entity_id": {
      "receiver.id": "snowflake-123",
      "receiver.name": "snowflake"
    },
    "destination_entity_type": "Snowflake",
    "destination_entity_id": {
      "receiver.id": "snowflake-456",
      "receiver.name": "snowflake"
    }
  }
]
```
___

## Entity Attributes
### Without Prefix
Entity update log will be sent together with an attribute, as configured in the entity definitions above.

**Defined events**

```yaml
events:
  entities:
    - type: KubernetesCluster
      action: update
      context: log
```

**Input resource attributes**

```
cluster.uid -> "cluster-123"
cluster.name -> "Cluster 123"
```

**Output log records**
```json
[
  {
    "entity_type": "KubernetesCluster",
    "entity_id": {
      "cluster.uid": "cluster-123"
    },
    "entity_attributes": {
      "cluster.name": "Cluster 123"
    }
  }
]
```
___
### With Prefix
Attribute will not be sent, because the prefix is not accepted when entity is not used as source or destination
entity in a relationship.

**Defined events**

```yaml
events:
  entities:
    - type: KubernetesCluster
      action: update
      context: log
```

**Input resource attributes**

```
cluster.uid -> "cluster-123"
src.cluster.name -> "Cluster 123"
```

**Output log records**
```json
[
  {
    "entity_type": "KubernetesCluster",
    "entity_id": {
      "cluster.uid": "cluster-123"
    }
  }
]
```
___

## Relationship Attributes
### Without Prefix
An attribute will be sent, because the relationship is defined with the attributes.

:warning: Relationship attributes are used only from unprefixed attributes.

**Defined events**

```yaml
events:
  relationships:
    - type: Has
      action: update
      context: log
      source_entity: KubernetesCluster
      destination_entity: KubernetesNamespace
      attributes: ["additional.attribute"]
```

**Input resource attributes**

```
cluster.uid -> "cluster-123"
namespace.name -> "namesapce-123"
additional.attribute -> "some value"
```

**Output log records**
```json
[
  {
    "relationship_type": "Has",
    "source_entity_type": "KubernetesCluster",
    "source_entity_id": {
      "cluster.uid": "cluster-123"
    },
    "destination_entity_type": "KubernetesNamespace",
    "destination_entity_id": {
      "cluster.uid": "cluster-123",
      "namespace.name": "namespace-123"
    },
    "relationship_attributes": {
      "additional.attribute": "some value"
    }
  }
]
```
___
### With Prefix and Entities
Attributes will be sent for:
- relationship, because it is defined in the `events.relationships` section.
- entities, because they are defined in the `schema.entities` section.

**Defined events**

```yaml
events:
  entities:
    - type: KubernetesCluster
      action: update
      context: log
  relationships:
    - type: CommunicatesWith
      action: update
      context: log
      source_entity: KubernetesCluster
      destination_entity: KubernetesCluster
```

**Input resource attributes**

```
src.cluster.uid -> "cluster-123"
dst.cluster.uid -> "cluster-456"
src.cluster.name -> "Cluster 123"
dst.cluster.name -> "Cluster 456"
additional.attribute -> "some value"
```

**Output log records**
```json
[
  {
    "entity_type": "KubernetesCluster",
    "entity_id": {
      "cluster.uid": "cluster-123"
    },
    "entity_attributes": {
      "cluster.name": "Cluster 123"
    }
  },
  {
    "entity_type": "KubernetesCluster",
    "entity_id": {
      "cluster.uid": "cluster-456"
    },
    "entity_attributes": {
      "cluster.name": "Cluster 456"
    }
  },
  {
    "relationship_type": "CommunicatesWith",
    "source_entity_type": "KubernetesCluster",
    "source_entity_id": {
      "cluster.uid": "cluster-123"
    },
    "destination_entity_type": "KubernetesCluster",
    "destination_entity_id": {
      "cluster.uid": "cluster-456"
    },
    "relationship_attributes": {
      "additional.attribute": "some value"
    }
  }
]
```
___
