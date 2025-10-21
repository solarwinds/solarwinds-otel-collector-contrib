# AGENTS.md

## Tools usage
### CircleCI
This is how you proceed if you want to get CircleCI build failure logs:
* Whenever you are provided with CircleCI link in following format "https://app.circleci.com/pipelines/github/solarwinds/solarwinds-otel-collector-contrib/121/workflows/6628c298-1814-47ca-a408-1210f82be382/jobs/248", you should use #tools_get_build_failure_logs to get the build failure.
* If you are provided with just PullRequest link, you should first get it directly based on branch: 
    * projectSlug: `gh/solarwinds/solarwinds-otel-collector-contrib` 
    * branch: actual branch of PR (can be extracted using #get_pull_request).
