# Custom Checkout Action

This action checks out a repository so your workflow can access its contents. It supports checking out a single branch, tag, or commit SHA, and is designed to work with pull request workflows, including **pull_request_target**.

## Features
- Checks out code from a pull request merge commit for secure review workflows.
- Supports standard branch/tag/commit checkout for other event types.
- Includes a security sanity check to ensure the checked-out code matches the expected pull request head SHA (see [actions/checkout#518](https://github.com/actions/checkout/issues/518)).
- Configurable fetch depth.

## Inputs
| Name         | Description                                                      | Required | Default |
|--------------|------------------------------------------------------------------|----------|---------|
| **fetch-depth**  | The number of commits to fetch. Only the latest by default.      | false    | `1`      |

## Usage
```yaml
- name: Checkout PR code
  uses: ./.github/actions/checkout
  with:
    fetch-depth: 1  # Optional, defaults to 1
```

## How It Works
- For **pull_request** and **pull_request_target** events, the action checks out the PR merge commit for accurate testing.
- For other events, the action performs a standard checkout.
- The action runs a sanity check to verify that the checked-out commit matches the PR head SHA for security.

## Security
This action includes a sanity check to help prevent malicious code injection by verifying that the checked-out commit matches the expected PR head SHA. For more details, see [actions/checkout#518](https://github.com/actions/checkout/issues/518).
