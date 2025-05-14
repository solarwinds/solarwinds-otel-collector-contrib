# Framework Concept & Motivation
Motivation for this framework was to safe as much code as possible. Repetitive parts of multiple scraper implementations were extracted and unified under one place to ensure its optimization and easy maintenance.

Framework concept introduces a way how to describe scraper structure declaratively. Only data collection specific parts needs to be implemented. Such scraper produces collection of OTEL metrics.  

Framework implementation utilizes Go concurrency and runs all configured components in parallel as much as possible.

# Components

Framework works from scraper level, which needs to be registered in receiver by factory. Scraper in general contains at least one scope emitter. Scope emitter contains at least one metric emitter. Each component is described in details below.

Receiver component structure

```
- Receiver
    - Scraper
       - Scope Emitter
           - Metric Emitter
           - Metric Emitter
       - Scope Emitter
           - Metric Emitter
    - Scraper
       - Scope Emitter
           - Metric Emitter
```

## Scraper Factory

Scraper factory creates an instance of scraper. Factory needs to implement [`types.ScraperFactory`](../../types/scrapers.go) interface to be registrable to receiver. Registration into receiver can be seen [here](#registration-in-swohostmetrics-receiver).

Created scraper instance needs to implement `scraperhelper.Scraper` interface. Framework provides generic function [`scraper.CreateScraper`](./scraper/factory.go) for scraper creation. This function can be used only if there is no need to introduce custom code before scraper creation. When custom code before scraper creation is required, classic implementation needs to be used ended by `scraperhelper.NewScraper` function. Both ways are showed [here](./example/factory.go).

## Scraper

From OTEL Collector perspective, scraper is a component, which produces metrics. Metrics produced by one scraper should be from one semantical area.

From framework perspective, scraper only defines its underlying components. Scraper declarative approach is explained [here](#declarative-approach).

Framework's scraper must implement interface [`scraper.Scraper`](./scraper/scraper.go). To be fully incorporated into framework processing there is a need to extended scraper struct by nested property of `scraper.Manager`.

```go
import ".../framework/scraper"

type OurScraper struct {
    scraper.Manager
}
```

Implementation of `scraper.Manager` interface is also provided by framework when struct is created:

```go
import ".../framework/scraper"

s := &OurScraper {
    Manager: scraper.NewScraperManager()
    ...
}
```

## Scope Emitters

Scope emitter produces scope metrics from OTEL Collector perspective.

From framework perspective scope emitters only manage metric emitters to be as concurent as possible. In case there is a need to have custom scope emitter, framework's declarative approach offers a allocator callback. Eventual custom scope emitter needs to implement [`scope.Emitter`](./scope/emitter.go) interface to be framework compliant. For more details take a look into [Declarative Approach](#declarative-approach) section.

## Metric Emitters

Metric emitter is a core component, which actually processes the most of work and contains all that specific code, which handles data gathering. It produces exactly one metric.

Metric emitter needs to implement [`metric.Emitter`](./metric/emitter.go) interface to be framework compliant.

There is an initiative for continuing scraping framework to metric receiver level and bellow. It would reduce amount of code needed for implementing metric emitter. It is covered by JIRA task [NH-76700](https://swicloud.atlassian.net/browse/NH-76700).

# Framework Utilization

From user perspective, there was an idea to introduce only declarative scraper configuration with metric emitters constructing callbacks. Followed by an introduction of nested property in scraper struct for scraping manager. And that's it. Scraping should be working in parallel through whole scraper tree. Example for such scraper is available [here](https://github.com/solarwinds-cloud/uams-otel-collector-plugin/tree/master/pkg/receiver/swohostmetricsreceiver/internal/scraper/framework/example).

## Registration in `swohostmetrics` receiver

When scraper is finished (can be tested separately via unit test), it must be registered in receiver to be ran in OTEL Collector.

Registration of [scraper factory](#scraper-factory) into `swohostmetrics` receiver is in [`scraperFactories` function](https://github.com/solarwinds-cloud/uams-otel-collector-plugin/blob/master/pkg/receiver/swohostmetricsreceiver/receiver.go#L24)

If scraper factory is implemented properly and scraper configuration is contained in OTEL collector config, scraper will start out of the box.

## Declarative Approach

Every scraper needs to provide declarative configuration about its underlying structure. The configuration consists of immediate collection of scope emitters. Scope emitters describes its own underlying structure as a collection of metric emitters, which belongs to scope emitter's scope.

Real working example from [exemplary scraper](https://github.com/solarwinds-cloud/uams-otel-collector-plugin/blob/master/pkg/receiver/swohostmetricsreceiver/internal/scraper/framework/example/scraper.go) follows

```go
// Scraper abilities. Declarative description of scopes and related metrics.
scraperDescription := &frameworkscraper.Descriptor{
    Name: ScraperName,
    ScopeDescriptors: map[string]scope.Descriptor{
        // Emits scope otelcol/swohostmetricsreceiver/exemplary-scraper/scope1
        scope1: {
            ScopeName: scope1,
            MetricDescriptors: map[string]metric.Descriptor{
                // Emits metric swo.exemplary-scraper.scope1.metric1.
                scope1metric1: {Create: NewMetricEmitterS1M1},
                // Emits metric swo.exemplary-scraper.scope1.metric2
                scope1metric2: {Create: NewMetricEmitterS1M2},
            },
        },
        // Emits scope otelcol/swohostmetricsreceiver/exemplary-scraper/scope1
        scope2: {
            ScopeName: scope2,
            MetricDescriptors: map[string]metric.Descriptor{
                // Emits metric swo.exemplary-scraper.scope2.metric1
                scope2metric1: {Create: NewMetricEmitterS2M1},
            },
        },
    },
}
```

## Currently Utilizing Solutions

Currently we have three scrapers implementing framework solution, so inspiration can be taken there.
- [Asset scraper](https://github.com/solarwinds-cloud/uams-otel-collector-plugin/tree/master/pkg/receiver/swohostmetricsreceiver/internal/scraper/assetscraper) producing two metrics
    - `swo.asset.installedsoftware`
    - `swo.asset.installedupdates`
- [Hardware Inventory scraper](https://github.com/solarwinds-cloud/uams-otel-collector-plugin/tree/master/pkg/receiver/swohostmetricsreceiver/internal/scraper/hardwareinventoryscraper) producing one metric
    - `swo.hardwareinventory.cpu`
- [Host Info scraper](https://github.com/solarwinds-cloud/uams-otel-collector-plugin/tree/master/pkg/receiver/swohostmetricsreceiver/internal/scraper/hostinfoscraper) producing three metrics
    - `swo.hostinfo.firewall` - Windows only
    - `swo.hostinfo.uptime`
    - `swo.hostinfo.user.lastLogged`
