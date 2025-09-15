# swok8sdiscovery Receiver

This receiver processes Kubernetes Containers and Services with a non-empty `external` attribute.

- For each matching Container, it publishes an event with the container name.
- For each matching Service, it publishes an event with the `external` attribute.
