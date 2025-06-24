# Manage Workflow Controllers

This guide explains how to manage Workflow Controllers. Workflow Controllers are responsible for orchestrating and triggering jobs in GitHub Actions workflows, especially for advanced scenarios like merge queues.

## Adding a Workflow to a Controller

To add a new workflow to a workflow controller, follow these steps:

1. Create or update the downstream workflow.  
   Ensure your workflow file (for example, `build-my-new-job.yml`) exists in `.github/workflows/` and is ready to be triggered by the controller. The workflow must use the **workflow_call** event in its `on` section:
   
   ```yaml
   on:
     workflow_call:
   ```

2. Add a filter for the workflow.  
	 Edit `.github/controller-filters.yaml` and add a new filter entry for your job. Use the naming convention `<job_name>-filter`, and specify the relevant file paths that should trigger this workflow.

   Example:
   ```yaml
   build-my-new-job-filter:
     - "cmd/my-new-job/*.go"
     - "cmd/my-new-job/Dockerfile"
     - "pkg/**"
     - "go.mod"
     - "go.sum"
   ```

3. Reference the workflow in the controller. 
   Edit the appropriate workflow controller file (for example, `workflow-controller-build-1.yml`). Add a new job that uses your workflow and is conditionally triggered by the filter you defined.

   Example:
   ```yaml
   build-my-new-job:
     needs: detect-changed-files
     if: ${{ contains(needs.detect-changed-files.outputs.files, 'build-my-new-job-filter') }}
     uses: kyma-project/test-infra/.github/workflows/build-my-new-job.yml@main
   ```

4. Set custom permissions if needed.  
   If the referenced workflow requires custom permissions, you must also set these permissions at the workflow controller level. GitHub Actions only allows jobs to use permissions that are granted at the top-level workflow (the controller). Add or update the `permissions` block in your controller YAML to include all permissions that downstream workflows need.

   > [!NOTE]
   > If you do not set the required permissions at the controller level, the downstream workflow will not have access to them, even if they are defined in the called workflow file.

   Example:
   ```yaml
   permissions:
     id-token: write
     contents: read
     issues: write  # Add any additional permissions required by the called workflow
   ```

   
## Related Information

- ADR-004: Adoption of Merge Queue
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [dorny/paths-filter Action](https://github.com/dorny/paths-filter)
