---
description: Guidelines for updating and reviewing the CHANGELOG.md file
applyTo: '**/CHANGELOG.md'
---

When updating the CHANGELOG.md file, please follow these guidelines:

There can be two main scenarios for updating the changelog:
1. **Releasing a new version**
2. **Making significant changes**
PRs without significant changes are not required to update the changelog (this is for PRs with chore in title, or similar non-functional changes)

You can tell the two apart easily. 
If new version appears in the changelog, it is a release update.
If new lines are added to vNext section, changelog is being updated for significant changes.

# Reviewing changelog

When reviewing a changelog update, ensure that:
- The version number follows semantic versioning (MAJOR.MINOR.PATCH).
- updates contain name of the component affected (if applicable).
- the type of change is clearly indicated (Added, Changed, Fixed, Deprecated, Removed, Security, BreakingChange).
- Each entry is concise yet descriptive enough to understand the change.

## Examples
```markdown
## vNext
- Added `swok8sdiscovery` receiver
```