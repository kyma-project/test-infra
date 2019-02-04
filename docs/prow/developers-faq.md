# Developers FAQ
Frequently Asked Questions (FAQ) a developer can have, regarding working with Prow as a Continuous Integration (CI) mechanism.

**Q: I would like to add a new component/item and have it operated by Prow**

**A:** Creating a CI pipeline is a two step process. First of all you need to define a ProwJob according to [this instruction](create-component-jobs.md). After that, you need to add the job in the release, from which you want it to be available. This requires some modifications to the created ProwJob, which can be found in this [document](create-release-jobs.md)

---
**Q: How to test changes I made to script definied in the test-infra repository?**

**A:** In order to do this we have 2 possible solutions:

1. Create a Pull Request (PR) with your changes and wait for the existing ProwJobs to verify your code

	> **NOTE**: This will work only if you modify existing code, and requires a PR for each consecutive change.

2. Add an `extra_refs` field to your ProwJob and work directly on Your branch. This will pull your chosen repository/branch into the job and execute the code from there:

	> **NOTE**: Remember to remove/change this after your code has been merged

```yaml
extra_refs:
  - org: aszecowka									# Your github user/organisation
    repo: test-infra								# Your github repository
    base_ref: dex-github							# Branch/tag/release to be used
    path_alias: github.com/kyma-project/test-infra 	# Location where to clone
```

---
**Q: How does the release process look like?**

**A:** We have created a document showcasing the whole process, please take a look [here](release-process.md)

---
**Q: My component is no longer needed, how do I remove it?**

**A:** In order to remove a component from Prow, follow these steps:
1. Delete **PreSubmit** and **PostSubmit** jobs for `master` branch. (Do not delete jobs for `release` branches)
2. Delete tests for these jobs.

> **NOTE**: If the component You have created is a part of a release *X*, You **cannot** just delete it, as it will be required in *X.y* (f.e a component in 0.6 that is deleted in 0.7 is still needed for 0.6.1)

---
**Q: The name of my component needs to change, what now?**

**A:** In the case of renaming a component, please follow these steps:
1. Change `name` value for presubmit and postsubmit jobs for `master` branch.
2. Change `run_if_changed` value in job_template for new component path.
3. Create new release job. For details, see the [official documentation](https://github.com/kyma-project/test-infra/blob/master/docs/prow/create-release-jobs.md).
>**NOTE**: Don't delete existing release jobs and tests for these jobs, as the component you're renaming in previous releases is still defined by its old name.
4. Make changes in tests.

