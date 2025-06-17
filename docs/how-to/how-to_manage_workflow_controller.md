# Manage Workflow Controllers

This guide explains how to manage Workflow Controllers. Workflow Controllers are responsible for orchestrating and triggering jobs in GitHub Actions workflows, especially for advanced scenarios like merge queues.

## Workflow Controller Overview

Workflow Controllers are special GitHub Actions workflows that:
- Detect changes in the repository.
- Use filters to determine which jobs should run.
- Trigger downstream workflows based on the detected changes.

They are essential for efficient CI/CD, especially when using features like merge queues.

## Key Files

- `.github/workflows/workflow-controller-*.yml`: Main workflow controller definitions (e.g., `workflow-controller-pull-requests.yml`, `workflow-controller-build-1.yml`, `workflow-controller-build-2.yml`).
- `.github/controller-filters.yaml`: Contains path-based filters for each job. These filters determine which jobs are triggered based on file changes.

## How It Works

1. **Change Detection**  
   Each workflow controller uses the [`dorny/paths-filter`](https://github.com/dorny/paths-filter) action to detect which files have changed in a PR or push.

2. **Filter Application**  
   The controller reads `.github/controller-filters.yaml` to map changed files to specific jobs. Each job has a corresponding filter (e.g., `build-automated-approver-filter`).

3. **Conditional Job Execution**  
   Jobs are triggered only if their filter matches the changed files. This is done using the `if:` condition in the workflow YAML.

4. **Downstream Workflow Triggers**  
   When a filter matches, the controller triggers the corresponding downstream workflow (e.g., `build-automated-approver.yml`).

## Example

A simplified example from `workflow-controller-build-1.yml`:

```yaml
jobs:
  detect-changed-files:
    ...
    - uses: dorny/paths-filter@v3
      id: pathFilters
      with:
        filters: .github/controller-filters.yaml

  build-automated-approver:
    needs: detect-changed-files
    if: ${{ contains(needs.detect-changed-files.outputs.files, 'build-automated-approver-filter') }}
    uses: kyma-project/test-infra/.github/workflows/build-automated-approver.yml@main
```

## Managing Filters

To add or update filters:
- Edit `.github/controller-filters.yaml`.
- Follow the naming convention: `<job_name>-filter`.
- Specify the relevant paths for each job.

Example filter entry:
```yaml
build-automated-approver-filter:
  - "cmd/external-plugins/automated-approver/*.go"
  - "cmd/external-plugins/automated-approver/Dockerfile"
  - "pkg/**"
  - "go.mod"
  - "go.sum"
```

## Adding a Workflow to a Controller

To add a new workflow to a workflow controller, follow these steps:

1. **Create or Update the Downstream Workflow**  
   Ensure your workflow file (e.g., `build-my-new-job.yml`) exists in `.github/workflows/` and is ready to be triggered by the controller. The workflow must use the `workflow_call` event in its `on` section:
   
   ```yaml
   on:
     workflow_call:
   ```

2. **Add a Filter for the Workflow**  
   Edit `.github/controller-filters.yaml` and add a new filter entry for your job. Use the naming convention `<job_name>-filter` and specify the relevant file paths that should trigger this workflow.

   Example:
   ```yaml
   build-my-new-job-filter:
     - "cmd/my-new-job/*.go"
     - "cmd/my-new-job/Dockerfile"
     - "pkg/**"
     - "go.mod"
     - "go.sum"
   ```

3. **Reference the Workflow in the Controller**  
   Edit the appropriate workflow controller file (e.g., `workflow-controller-build-1.yml`). Add a new job that uses your workflow and is conditionally triggered by the filter you defined.

   Example:
   ```yaml
   build-my-new-job:
     needs: detect-changed-files
     if: ${{ contains(needs.detect-changed-files.outputs.files, 'build-my-new-job-filter') }}
     uses: kyma-project/test-infra/.github/workflows/build-my-new-job.yml@main
   ```

4. **Set Custom Permissions (if needed)**  
   If the referenced workflow requires custom permissions, you must also set these permissions at the workflow controller level. GitHub Actions only allows jobs to use permissions that are granted at the top-level workflow (the controller). Add or update the `permissions` block in your controller YAML to include all permissions needed by downstream workflows.

   Example:
   ```yaml
   permissions:
     id-token: write
     contents: read
     issues: write  # Add any additional permissions required by the called workflow
   ```

   > **Note:** If you do not set the required permissions at the controller level, the downstream workflow will not have access to them, even if they are defined in the called workflow file.
   
## References

- ADR-004: Adoption of Merge Queue
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [dorny/paths-filter Action](https://github.com/dorny/paths-filter)

