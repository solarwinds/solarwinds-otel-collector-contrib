# Contributing

Thank you for your interest in contributing to SolarWinds OpenTelemetry Collector Contrib!

Currently, this project is being developed by the SolarWinds team. 
We are actively working on stabilizing its foundations and defining contribution guidelines.
Once ready, we will update the `CONTRIBUTING.md` file with clear instructions.

## Before Adding a New Component

Before proposing or implementing a new OpenTelemetry component, please consider the following:

### Evaluate Necessity
- **Is a new component truly needed?** Many telemetry transformation requirements can be solved using existing components, especially with processors like [transform](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/transformprocessor), [filter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor), or [metricstransform](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/metricstransformprocessor).
- Review the [current components list](README.md#components) and the [OpenTelemetry Collector Contrib repository](https://github.com/open-telemetry/opentelemetry-collector-contrib) to avoid duplication.

### Contact Us First
- For SolarWinds internal contributors: Reach out in the `#swotelcol-platform` Slack channel to discuss your proposal.
- For external contributors: Open an issue describing your use case and proposed component before investing significant time in development.

Before any code is written contact us providing the following information:
* Who's the sponsor for your component. A sponsor is an approver or maintainer who will be the official reviewer of the code and a code owner for the component.
* Some information about your component, such as the reasoning behind it, use-cases, telemetry data types supported, and anything else you think is relevant for us.
* The configuration options your component will accept. This will give us a better understanding of what it does, and how it may be implemented.