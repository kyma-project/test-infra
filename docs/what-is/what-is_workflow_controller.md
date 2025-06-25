# Workflow Controller

Workflow Controller is a GitHub Actions workflow that orchestrates and triggers downstream workflows based on repository changes. It implements advanced CI/CD logic, such as merge queues and selective job execution, by applying path-based filters to changed files. It was created due to a lack of filtering capabilities in GitHub Actions workflows for merge queues.

## Architecture and Key Files

- Workflow Controller definitions: `.github/workflows/workflow-controller-*.yml` (for example, `workflow-controller-pull-requests.yml`, `workflow-controller-build-1.yml`).
- Filter configuration: `.github/controller-filters.yaml` — defines path-based filters for each job or workflow.
- Downstream workflows: `.github/workflows/*.yml` — must support `workflow_call` triggers.

## Workflow Controller Mechanics

1. Change detection: Uses [`dorny/paths-filter`](https://github.com/dorny/paths-filter) to detect changed files in a PR or push event.

2. Filter application: Loads filters from `.github/controller-filters.yaml`. Each filter (for example, `build-automated-approver-filter`) specifies file globs that, when matched, should trigger a job.

3. Conditional job execution: Each job in the controller workflow uses an `if:` condition to check if its filter matched any changed files.

4. Downstream workflow triggering: If a filter matches, the controller triggers the corresponding downstream workflow using the **uses** keyword and referencing the workflow file.

## Examples

See an example of the controller workflow:

```yaml
jobs:
  detect-changed-files:
    runs-on: ubuntu-latest
    steps:
      - uses: dorny/paths-filter@v3
        id: pathFilters
        with:
          filters: .github/controller-filters.yaml

  build-automated-approver:
    needs: detect-changed-files
    if: ${{ contains(needs.detect-changed-files.outputs.files, 'build-automated-approver-filter') }}
    uses: kyma-project/test-infra/.github/workflows/build-automated-approver.yml@main
```

See an example of filter configuration in the following `.github/controller-filters.yaml` file:

```yaml
build-automated-approver-filter:
  - "cmd/external-plugins/automated-approver/*.go"
  - "cmd/external-plugins/automated-approver/Dockerfile"
  - "pkg/**"
  - "go.mod"
  - "go.sum"
```

## Permissions

If downstream workflows require custom permissions, you must set them at the controller level in the `permissions` block. Permissions defined only in the downstream workflow are not sufficient.

## Related Information

- [Manage Workflow Controllers](../how-to/how-to_manage_workflow_controller.md)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [dorny/paths-filter Action](https://github.com/dorny/paths-filter)
- ADR-004: Adoption of Merge Queue

