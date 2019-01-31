# Developers FAQ
Frequently Asked Questions (FAQ) a developer can have, regarding working with Prow as a Continuous Integration (CI) mechanism.

**Q: I would like to add a new component/item and have it operated by Prow**

**A:** Creating a CI pipeline for a new component requires a new ProwJob. Please see [this document](create-component-jobs.md) for further information

**Q: I need my component to be a part of a release, how to do it?**

**A:** A release job is a special kind of ProwJob that uses a specific release branch. Please see [this document](create-release-jobs.md) for further information

**Q: I would like to use Prow to check something on my fork, how should I do it?**

**A:** In order to do this we have 2 possible solutions:

1. Create a Pull Request (PR) with your changes and wait for the existing ProwJobs to verify your code

	> **NOTE**: This will work only if you modify existing code, and requires a PR for each consecutive change.

2. Add an `extra_refs` field to your ProwJob and work directly on Your branch:

	> **NOTE**: Remember to remove/change this after your code has been merged

```yaml
extra_refs:
  - org: aszecowka # TODO only temporary solution
    repo: test-infra
    base_ref: dex-github
    path_alias: github.com/kyma-project/test-infra
```
