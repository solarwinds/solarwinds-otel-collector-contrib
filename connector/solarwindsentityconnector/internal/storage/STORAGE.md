# Storage

The storage is used to track relationship lifetime. If the relationship was not seen for longer than the set interval, an event is sent to the SolarWinds Observability SaaS to delete the relationship. The storage is also used to track entity lifetime. If the entity was not seen for longer than the set interval, an event is sent to the SolarWinds Observability SaaS to set the entity as unknown.

The connector can be configured to use internal storage in configuration file
We differentiate between two types of storage:
- internal storage: used to store the state of entities and relationships in memory
