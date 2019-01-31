# Developers FAQ
Frequently Asked Questions (FAQ) a developer can have, regarding working with Prow as a Continuous Integration (CI) mechanism.

**Q: I would like to add a new component/item and have it operated by Prow**

**A:** Creating a CI pipeline for a new component requires a new ProwJob. Please see [this document](create-component-jobs.md) for further information

---
**Q: I need my component to be a part of a release, how to do it?**

**A:** A release job is a special kind of ProwJob that uses a specific release branch. Please see [this document](create-release-jobs.md) for further information

---
**Q: I would like to use Prow to check something on my fork, how should I do it?**

**A:** In order to do this we have 2 possible solutions:

1. Create a Pull Request (PR) with your changes and wait for the existing ProwJobs to verify your code

	> **NOTE**: This will work only if you modify existing code, and requires a PR for each consecutive change.

2. Add an `extra_refs` field to your ProwJob and work directly on Your branch. This will pull your chosen repository/branch into the job and execute the code from there:

	> **NOTE**: Remember to remove/change this after your code has been merged

```yaml
extra_refs:
  - org: aszecowka # TODO only temporary solution
    repo: test-infra
    base_ref: dex-github
    path_alias: github.com/kyma-project/test-infra
```

---
**Q: How does the release process look like?**

**A:** We have created a document showcasing the whole process, please take a look [here](release-process.md)

---
**Q: My component is no longer needed, how do I remove it?**

**A:** In order to remove a component from Prow, we need to backtrack and remove everything we have created in [this document](create-component-jobs.md). 

> **NOTE**: If the component You have created is a part of a release *X*, You **cannot** just delete it, as it will be required in *X.y* (f.e a component in 0.6 that is deleted in 0.7 is still needed for 0.6.1)

In such a situation it is required to remove the **PreSubmit** and **PostSubmit** ProwJob triggers for the **master branch**, while leaving the triggers for the **release branch only**

---
**Q: The name of my component needs to change, what now?**

**A:** In the case of renaming a component, please take a look [here](create-component-jobs.md#Rename-a-component)
