#### Description
Provide description of the changes made.

**Tracking Issue:** https://github.com/solarwinds/solarwinds-otel-collector-contrib/issues/XXXX

#### Testing
Describe what testing was performed and which tests were added.

---

### Contribution Checklist

Please ensure your PR meets the following requirements (see [CONTRIBUTING.md](../CONTRIBUTING.md)):

**General:**
- [ ] **CHANGELOG updated** - Added entry under `## vNext` (chore PRs are a possible exeption)
- [ ] **Tests included** - Unit tests, generated tests (`mdatagen`), and/or integration tests
- [ ] **Documentation updated** - README.md with config examples and behavior description

**For new components:**
- [ ] **Codeowners** - Approver/maintainer assigned
- [ ] **Use case documented** - Reasoning, telemetry types, configuration options described in issue
- [ ] `mdatagen` has been used to autogenerate tests, update README.md and more (see [CONTRIBUTING.md](../CONTRIBUTING.md)):
- [ ] README.md contains complete configuration documentation
- [ ] **Telemetry names registered** - New metrics registered (if applicable)
> [!NOTE]  
> New components need to be added into the release (one or multiple distributions in our private releases repository).  
That can be done only after the new component has been merged & released.
