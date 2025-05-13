# Development Guidelines

## Naming

### Branches

Name of the branches created within this repository should follow the following format:

`{type}/#{issue-number}-{description-of-changes}`

_Example:_ `chore/#12-update-dependency-version`

As for individual parts of the branch name:

- _type_ should reflect the types described in [Commits](#commits) section
- _issue-number_ should refer to related [GitHub Issue](https://github.com/solarwinds/solarwinds-otel-collector-contrib/issues) number. If such issue does not exist (i.e. for immediate fixes) this part starting with hash should be omitted
- _description-of-changes_ should provide meaningful description of changes the branch holds, with words separated by dashes (i.e. title of related [GitHub Issue](https://github.com/solarwinds/solarwinds-otel-collector-contrib/issues), or its respective sub-part)

### Commits

Commit naming in general should follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) conventions. Please refer to related [documentation](https://www.conventionalcommits.org/en/v1.0.0/) for the specification with examples.

For _type_ part of the commit message, utilize the full palette of suggested types based of [Angular conventions](https://github.com/angular/angular/blob/22b96b9/CONTRIBUTING.md#type). 
For feature-altering commits utilize `feature` over `feat` type name.

_Example:_ `feature: add collector_name parameter to solarwindsextension`

> [!WARNING]  
> Please pay extra attention to commit naming when creating squash merge commits to protected branches.

### Pull Requests

Pull requests should reflect [branch name](#branches) they are created from. If the branch name is not descriptive enough, the pull request title should be adjusted to be as descriptive as possible. The format should be following:

`{type}: #{issue number} {Description of Changes}`

_Example:_

- branch name: `chore/#12-update-dependency-version`
- pull request title: `chore: #12 Update Dependency Version`

### Code

Naming within code should in general follow principles laid out by [Effective Go](https://go.dev/doc/effective_go) with having [OpenTelemetry naming conventions](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/coding-guidelines.md#naming-convention) in mind.

## Releasing a new version

To release a new version of SolarWinds OpenTelemetry Collector:

1. Make sure that the `## vNext` section in [CHANGELOG.md](../CHANGELOG.md) contains all changes.
2. Run

    ```shell
    make prepare-release version=0.113.0
    ```

    replacing `0.113.0` with an actual version to release.

3. Verify the generated changes, prepare a PR and merge it.
4. This will trigger a build pipeline that will release a new version of the collector.
