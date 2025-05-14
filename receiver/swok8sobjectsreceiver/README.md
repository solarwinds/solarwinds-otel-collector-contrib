# SWO Kubernetes Objects Receiver

The SWO Kubernetes Objects Receiver (swok8sobjects) is based on the existing [k8sobjectsreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/k8sobjectsreceiver) from OpenTelemetry. In addition to the original functionality, it adds attributes to the log messages indicating which part of the manifest was changed (`sw.metadata.changed`, `sw.status.changed`, `sw.spec.changed`). 

Additionally, this receiver can be configured to use storage, allowing it to persist data across collector restarts.

The following is an example configuration:

```yaml
swok8sobjects:
  auth_type: serviceAccount
  storage: file_storage/manifests
  objects:
    - name: pods
      mode: pull
      label_selector: environment in (production),tier in (frontend)
      field_selector: status.phase=Running
      interval: 15m
    - name: events
      mode: watch
      group: events.k8s.io
      namespaces: [default]
```
