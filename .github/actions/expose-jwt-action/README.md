# View GitHub JWT Token

This action takes an optional input for the `aud` value and outputs a GitHub signed JWT. 

## Inputs

## `audience`

The audience field in the JWT

Default: ``

## Outputs

## `jwt`

The JSON Web Token signed by GitHub Actions

## Example usage

uses: ./.github/actions/expose-jwt-action@v0.1.5
with:
  audience: 'sts.amazonaws.com'