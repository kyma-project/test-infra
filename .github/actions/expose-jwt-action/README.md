# Expose GitHub JSON Web Token Action

This action takes an optional input for the **audience** value and outputs a GitHub-signed JSON Web Token (JWT).

## Inputs

## **audience**

The **audience** field in the JWT.

Default: ``

## Outputs

## `jwt`

The JWT signed by GitHub Action.

## Example Usage
```yaml
# To use this repository's private action,
# you must check out the repository
- uses: actions/checkout@v4 
  name: Checkout
# Install Node.js and needed dependencies
- uses: ./.github/actions/expose-jwt-action/install
  name: Install expose-jwt-action
# This action is used to expose the JWT token from the OIDC provider and set is as an output and an environment variable
- uses: ./.github/actions/expose-jwt-action
  name: Expose JWT token
  with:
    audience: 'sts.amazonaws.com'
```