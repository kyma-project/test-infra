# OIDC Token Verifier

The OIDC Token Verifier is a command-line tool designed to validate the OIDC token and its claims values. It is primarily used in the
oci-image-builder pipeline to authenticate and ensure the integrity of the token passed to the pipeline.

At present, the tool supports only github.com OIDC identity provider and the RS256 algorithm for verifying the token signature.

## How to use

- authorisation env var
- expected value of the job workflow ref

To use the OIDC Token Verifier as a command-line tool, you need to pass the raw OIDC token and the issuer information as arguments. The tool
will then verify the token and extract the claims.

Here is an example of how to use the OIDC Token Verifier from the command line:

```bash
oidc-token-verifier --token "your-oidc-token"
```

Please replace `"your-oidc-token"` with your actual OIDC token.

## How it works

- the oidc discovery
- the token and claims verification
- hardcoded trusted issuer and workflow, link to issue

The OIDC Token Verifier works by first using the provided verifier to verify the token's signature and expiration time. If the verification
is successful, it then extracts the claims from the token.

The tool is designed to be used in the oci-image-builder pipeline, where it helps to ensure the integrity and authenticity of the tokens
used in the pipeline. By verifying the tokens, it helps to prevent unauthorized access and potential security risks.