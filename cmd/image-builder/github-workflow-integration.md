# Image Builder GitHub Workflow Integration

The Image Builder solution integrates with GitHub workflows and uses an Azure DevOps pipeline to automate the process of building OCI
images. It leverages a signed JWT format in which an OIDC token from GitHub's OIDC identity provider is passed. This token is used for
secure and authorized passing of information about the workflow and the image to build to the `oci-image-builder` pipeline. The build
process is executed in an Azure DevOps pipeline, providing a slc-29 compliant infrastructure for the building of OCI images.

## Process Flow

1. **Obtaining the OIDC Token**: The process begins with GitHub's OIDC identity provider issuing an OIDC token. This token is used to
   securely pass information about the workflow.

2. **Passing the OIDC Token**: The OIDC token is then passed as a parameter to the `oci-image-builder` pipeline in a signed JWT format. This
   ensures the secure and authorized passing of information about the workflow.

3. **Validating the OIDC Token**: The `oci-image-builder` pipeline, running in Azure DevOps, validates the OIDC token against the GitHub's
   OIDC identity provider. This step ensures that the token is valid and has not been tampered with.

4. **Using the OIDC Token**: The `oci-image-builder` pipeline uses the information from the OIDC token to clone the appropriate source code
   for the building of the OCI image. Additionally, it uses the information from the OIDC token to set the appropriate parameters for the
   build and to decide whether the OCI image should be signed or not.

5. **Building the OCI Image**: Once the source code is cloned and the build parameters are set, the `oci-image-builder` pipeline proceeds to
   build the OCI image. This process involves compiling the source code and packaging it into an OCI image.

6. **Signing the OCI Image**: If the build was triggered by a push event in GitHub and the OIDC token indicates that the image should be
   signed, the `oci-image-builder` pipeline uses the `signify` service to sign the OCI image. This step ensures the integrity and
   authenticity of the OCI image.

7. **Pushing the OCI Image**: After the OCI image is built and optionally signed, it is then pushed to a specified OCI registry.

## GitHub OIDC Identity Token Claims

The OIDC token issued by GitHub's OIDC identity provider contains several claims that are crucial for the `oci-image-builder` pipeline.
These claims are used to identify the workflow triggering the build pipeline and to clone the appropriate version of the source code. This
is essential for slc-29 compliance, as it ensures that the exact version of the code that was tested in the PR or for which the push was
merged is built.

### Workflow Identification Claims

The OIDC token contains claims that can be used to identify the workflow that triggered the build pipeline. These include:

- `iss`: The issuer of the token. This is always `https://token.actions.githubusercontent.com`.
- `iat`: The time at which the token was issued.
- `exp`: The time at which the token expires.
- `jti`: A unique identifier for the token.
- `nbf`: The time before which the token must not be accepted for processing.
- `kid`: The key ID of the key used to sign the token.
- `alg`: The algorithm used to sign the token.
- `run_id`: The ID of the workflow run.
- `run_number`: The number of the workflow run.
- `actor`: The login of the user who initiated the workflow run.
- `event_name`: The name of the event that triggered the workflow run.
- `workflow`: The name of the workflow that triggered the workflow run.
- `workflow_ref`: The git ref associated with the workflow file.
- `repository`: The repository where the workflow run occurred.
- `repository_owner`: The owner of the repository where the workflow run occurred.

### Source Code Cloning Claims

The OIDC token also contains claims that can be used to clone the appropriate version of the source code:

- `repository`: The repository where the workflow run occurred.
- `repository_owner`: The owner of the repository where the workflow run occurred.
- `event_name`: The name of the event that triggered the workflow run.
- `ref`: The git ref associated with the workflow run. This can be used to checkout the correct branch, tag, or commit.
- `ref_type`: The type of git ref associated with the workflow run. This can be used to determine whether the ref is a branch, tag, or
  commit.
- `base_ref`: The base git ref associated with the workflow run. This can be used to determine the base branch for a pull request.
- `head_ref`: The head git ref associated with the workflow run. This can be used to determine the head branch for a pull request.

These claims ensure that the `oci-image-builder` pipeline builds the exact version of the code that was provided in the PR or merged to the
branch, adhering to slc-29 compliance.

## Block Diagram

![image-builder-block-diagram](documentation_assets/image-builder-block-diagram.png)

## Activity Diagram

![image-builder-activity-diagram](documentation_assets/image-builder-activity-diagram.png)

## Conclusion

The Image Builder solution, with its seamless integration with GitHub workflows and Azure DevOps pipeline, offers developers a robust and
secure method to incorporate the building of OCI images into their workflows. By leveraging a signed JWT format in which an OIDC token from
GitHub's OIDC identity provider is passed, it ensures the secure and authorized passing of information about the workflow and the image to
build. The entire build process adheres to slc-29 compliance, providing a reliable infrastructure for the building of OCI images.
