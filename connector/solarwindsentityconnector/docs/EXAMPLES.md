# Examples
This document provides examples of how to configure the SolarWinds Entity Connector correctly to prevent from inferring
unexpected entities or relationships. 

Each example includes
- the defined events (what entity type/relationship should be inferred on what conditions),
- the input resource attributes (what attributes are in the incoming telemetry resource),
- the expected output log records (what will be sent to the SolarWinds Observability).

#### Table of Contents
- [Entity ID Attributes](#entity)
  - [Entity Without Prefix](#entity-without-prefix)
  - [Entity With Prefix](#entity-with-prefix)
  - [Two Entities Without Prefix](#two-entities-without-prefix)
  - [Two entities with the same set of IDs :warning:](#two-entities-with-the-same-set-of-ids-warning)
  - [Two entities with the same set of IDs with conditions](#two-entities-with-the-same-set-of-ids-with-conditions)

Assume following entities definitions with simplified attribute names and source/destination prefix.

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

## Entity ID Attributes

### Entity Without Prefix
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


### Entity With Prefix
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

### Two Entities Without Prefix
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

### Two entities with the same set of IDs :warning:
Two entities will be inferred, because conditions are true and all ID attributes for both entities are found, even if
it is not correct. To solve this issue, use `conditions` to filter out entities that should not be inferred.

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

### Two entities with the same set of IDs with conditions
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
### Different-type Relationship Without Prefix, Creating Entities
Relationship without prefixed attributes can be used to infer relationships between entities:
- of different types
- of different entity IDs sets.

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

### Different-type Relationship Without Prefix, Without entities
Example of the same scenarios as above, but without entities defined in the events section.

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

### Different-type Relationship with Prefixes, creating valid Entities
Relationship with prefixed attributes can be used to infer relationships between entities:
- of different types
  - with or without same set of IDs, (this example shows with same set of IDs = Snowflake -> DockerDaemon)
- of the same type.

Together with inferring entities, from prefixed attributes. However, in this scenario, we present case where
conditions for DockerDaemon entity are not met, so only Snowflake is not inferred (no conditions means *true*).

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


### Same-type Relationship with Prefixes, creating Entities
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

## Entity Attributes
### Entity with attributes
Entity update log will be sent together with an attribute.

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
### Entity with prefixed attributes
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
### Relationship with unprefixed attributes
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
      "cluster.uid": "namespace-456"
    },
    "relationship_attributes": {
      "additional.attribute": "some value"
    }
  }
]
```
___
### Relationship And Entities with prefixed attributes
Attributes will be sent for:
- relationship, because it is defined with the `events.relationship` section.
- for entities, because they are defined in the `schema.entities` section.

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



## Notes
- Resource Attributes are map so no duplicate entry is expected.
- `EXAMPLES.md` shows the scenarios simplified, for supported syntax see [README.md](../README.md)
or [OUTPUT.md](OUTPUT.md) to see actual output of the SolarWinds Entity Connector.
- `conditions: ["true"]` is the same as not defining conditions at all.