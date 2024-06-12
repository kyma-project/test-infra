# OIDC Token Verifier

The OIDC Token Verifier is a command-line tool designed to verify and extract claims from an OIDC token. It is primarily used in the
oci-image-builder pipeline to authenticate and ensure the integrity of the tokens used in the pipeline. The tool uses a provided verifier to
verify the token's signature and expiration time, and then extracts the claims from the token.

## How to use

To use the OIDC Token Verifier as a command-line tool, you need to pass the raw OIDC token and the issuer information as arguments. The tool
will then verify the token and extract the claims.

Here is an example of how to use the OIDC Token Verifier from the command line:

```bash
oidc-token-verifier --token "your-oidc-token" --issuer "https://your-issuer-url.com"
```

Please replace `"your-oidc-token"` and `"https://your-issuer-url.com"` with your actual OIDC token and issuer URL.

The tool will output the claims in a human-readable format, or it can be configured to output in a machine-readable format such as JSON.

## How it works

The OIDC Token Verifier works by first using the provided verifier to verify the token's signature and expiration time. If the verification
is successful, it then extracts the claims from the token.

The tool is designed to be used in the oci-image-builder pipeline, where it helps to ensure the integrity and authenticity of the tokens
used in the pipeline. By verifying the tokens, it helps to prevent unauthorized access and potential security risks.