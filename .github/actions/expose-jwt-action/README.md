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
- name: Checkout
  uses: actions/checkout@v4
- uses: actions/setup-node@v4
  with:
    node-version: 20
- run: cd .github/actions/expose-jwt-action
- run: npm init -y
- run: npm install @actions/core
- run: npm install @actions/github
- uses: ./.github/actions/expose-jwt-action
  name: Get JWT token
  with:
    audience: 'https://github.com/github'
```