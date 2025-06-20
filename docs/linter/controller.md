# Controller Linter Guide

This document describes the design and implementation of linter for the GitHub Actions **controller** (`controller.yml`) and the reusable workflows it triggers. The goal is to check at an early stage whether the jobs defined in it will run as expected depending on which files have been modified in the PR.

---

## Table of Contents

1. [Introduction](#introduction)
2. [Desired Controller Pattern](#desired-controller-pattern)
3. [Validation Rules](#validation-rules)
4. [Implementation Approach](#implementation-approach)
5. [Example Rules in Practice](#example-rules-in-practice)
6. [Conclusion](#conclusion)

---

## Introduction

Since the **merge_group** event does not support filtering by default which files have been changed, we decided to use `dorny/paths-filter@v3`. Using this tool gives us this possibility. We also want to have control over which reusable workflows are triggered in case of the `pull_request_target` event and which ones are triggered in case of the `merge_group` event. For this reason, we decided to create a controller that will control this. To prevent configuration drift and ensure that:

* Job keys match file names and `if` conditions
* Event triggers are correctly defined
* Filters cover exactly the set of jobs
* Limits on the number of reusable workflows are respected

the proposal is to implement a **Strategy-based linter** in Go. Each **Rule** encapsulates a single validation check against:

* `controller.yml` structure
* Local reusable workflows under `.github/workflows`
* The filter configuration in `controller-filters.yaml`

By running these rules as part of CI, the team can catch mistakes early.

---

## Desired Controller Pattern

The controller (`.github/workflows/controller.yml`) must follow these conventions:

* **Event Triggers** (`on`):

  ```yaml
  on:
    pull_request_target:
      types: [opened, edited, synchronize, reopened, ready_for_review]
    merge_group:
      types: [checks_requested]
  ```

* **Dispatch Job** (`run-workflows`):

  ```yaml
  jobs:
    run-workflows:
      runs-on: [self-hosted, solinas]
      outputs:
        files: ${{ steps.pathFilters.outputs.changes }}
      steps:
        - name: Checkout PR merge commit
          uses: actions/checkout@v4
        - uses: dorny/paths-filter@v3
          id: pathFilters
          with:
            filters: .github/controller-filters.yaml
  ```

* **Downstream Jobs**: each job key → reusable workflow:

    * The **job key** (e.g. `unit-test-go`) must match:

        * The filename in `uses: "./.github/workflows/unit-test-go.yml"`
        * The filter name in `if: ${{ contains(needs.run-workflows.outputs.files, 'unit-test-go-paths') }}`

---

## Validation Rules

Each of the following rules implements one aspect of the desired pattern:

1. **OnTriggersRule**

    * Verifies presence and exact list of `pull_request_target.types` and `merge_group.types`.

2. **UsesMatchRule**

    * Ensures each job’s `uses` path basename equals `<jobKey>.yml`.

3. **NeedsRunWorkflowsRule**

    * Checks that every job (except `run-workflows`) has `needs: run-workflows`.

4. **IfContainsFilesRule**

    * Validates each job’s `if` contains `contains(needs.run-workflows.outputs.files, '<jobKey>-paths')`.

5. **RunWorkflowsStructureRule**

    * Confirms `run-workflows` has the correct `runs-on`, `outputs.files`, and first two `steps.uses`.

6. **FilterKeysRule**

    * Loads `.github/controller-filters.yaml` and ensures:

        * It defines at least one entry per job.
        * Every key `<jobKey>-paths` corresponds to an actual job.

7. **WorkflowCallEventRule**

    * For each local reusable workflow (under `.github/workflows`), checks presence of `on.workflow_call`.

8. **NestedReusableCountRule**

    * Parses each local reusable workflow’s `jobs` section, collects nested `uses:` references,
    * Ensures total distinct reusable workflows (controller + nested) ≤ 20.

---

## Implementation Approach

The proposal is to use the **Strategy Pattern**:

* **Rule** interface:

  ```go
  type Rule interface {
    Name() string
    Validate(ctrl *Controller) []error
  }
  ```
* Each rules and validator in file under `pkg/linter/controller/rule.go`.
* **Parser** (`pkg/linter/controller/parser.go`) unmarshals `controller.yml` and `controller-filters.yaml`.

Example registration in `cmd/linter/controller/main.go`:

```go
rules := []Rule{
  &OnTriggersRule{},
  &UsesMatchRule{},
  &NeedsRunWorkflowsRule{},
  &IfContainsFilesRule{},
  &RunWorkflowsStructureRule{},
  &FilterKeysRule{},
  &WorkflowCallEventRule{},
  &MaxReusableJobsRule{},
  &NestedReusableCountRule{},
}
errs := ValidateController(ctrl, rules)
```

---

## Example Rules in Practice

### Pattern Enforcement

For job key `unit-test-go`:

```yaml
jobs:
  unit-test-go:
    uses: "./.github/workflows/unit-test-go.yml"
    needs: run-workflows
    if: ${{ contains(needs.run-workflows.outputs.files, 'unit-test-go-paths') }}
```

* `UsesMatchRule` ✔︎ ensures basename ends with `unit-test-go.yml`.
* `IfContainsFilesRule` ✔︎ ensures `if` uses the correct filter key.

### Filter File Two‑Way Matching

Given `controller-filters.yaml`:

```yaml
unit-test-go-paths:
  - '.github/workflows/unit-test-go.yml'
  - '**/*.go'
another-job-paths:
  - '...'
```

* `FilterKeysRule` must verify:

    * All keys `unit-test-go-paths`, `another-job-paths` match jobs `unit-test-go` and `another-job`.

---

## Conclusion

By codifying these rules with the **Strategy Pattern**, we achieve:

* **Automated validation** of controller structure and reusable workflows.
* **Early feedback** in CI for misconfigurations.
* **Clear documentation** of repository conventions.

As new rules emerge, simply add new `Rule` implementations without modifying existing logic. This approach provides us with code that is open to extensions and closed to modifications.

