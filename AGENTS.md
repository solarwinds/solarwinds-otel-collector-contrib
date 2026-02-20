# AGENTS.md

## Code Conventions

- Branch naming: `{type}/#{issue-number}-{description-of-changes}`
- Follow [development-guidelines.md](docs/development-guidelines.md) for all naming conventions

## Testing Conventions

- **Test file placement:** Each test file must mirror its source file in name and scope. For example, tests for functions in `log_constructor.go` belong in `log_constructor_test.go`. Do not create standalone test files for individual features — add tests to the file that tests the same SUT.

## Code Style

- **Comments must reflect current state:** Comments must be valid for the entire future lifetime of the code, ignoring PR context. Never use temporal markers such as `new:`, `existing:`, `// New `, `// Added ` in comments — these become incorrect the moment the PR is merged.

- **Prefer clean readable code over inline comments:** Write inline comments only when recording WHY, not WHAT. If a comment merely restates what the code does, delete it. Reserve inline comments for non-obvious constraints, trade-offs, or gotchas that cannot be expressed through naming alone.

## Tools usage
### CircleCI
This is how you proceed if you want to get CircleCI build failure logs:
* Whenever you are provided with CircleCI link in following format "https://app.circleci.com/pipelines/github/solarwinds/solarwinds-otel-collector-contrib/121/workflows/6628c298-1814-47ca-a408-1210f82be382/jobs/248", you should use #tools_get_build_failure_logs to get the build failure.
* If you are provided with just PullRequest link, you should first get it directly based on branch: 
    * projectSlug: `gh/solarwinds/solarwinds-otel-collector-contrib` 
    * branch: actual branch of PR (can be extracted using #get_pull_request).
