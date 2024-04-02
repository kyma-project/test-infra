# Image Builder GitHub Workflow Integration

The Image Builder solution integrates with GitHub workflows and uses an Azure DevOps pipeline to run the process of building OCI
images. It leverages a signed JWT format in which an OIDC token from GitHub's OIDC identity provider is passed. This token is used for
secure and authorized passing of information about the workflow and the image to build to the `oci-image-builder` pipeline. The build
process is executed in an Azure DevOps pipeline, providing an SLC-29-compliant infrastructure for building OCI images.

## Process Flow

1. **Trigger workflow**: The user or automation triggers a GitHub workflow.

2. **Obtaining the OIDC Token**: The workflow uses GitHub action to call GitHub's OIDC identity provider issuing an OIDC token.
   This token is used to securely pass information about the workflow.

3. **Trigger oci-image-builder pipeline**: GitHub action running Image Builder client uses a Personal Access Token (PAT) to authenticate
   with the Azure DevOps (ADO) API and trigger the `oci-image-builder` pipeline. The OIDC token and additional parameters required by
   the `oci-image-builder` pipeline are passed as parameters to the pipeline. These parameters' values are collected from data defined by
   the user and GitHub OIDC identity provider. The OIDC token is used for secure and authorized passing of information about the workflow
   and the image to build.

4. **Validating the OIDC Token**: The `oci-image-builder` pipeline, running in Azure DevOps, detects a call originating from GitHub and
   validates the OIDC token against GitHub's
   OIDC identity provider. This step ensures that the token is valid and has not been tampered with.

5. **OCI Image build preparation**: The `oci-image-builder` pipeline uses the information from the OIDC token to clone the appropriate
   source code for the building of the OCI image. It uses the information from the OIDC token and user-defined parameters to
   set the appropriate parameters for the image build and signing. The data from the OIDC token is used to decide whether the OCI image should
   be signed or not.

6. **Building the OCI Image**: Once the source code is cloned and the build parameters are set, the `oci-image-builder` pipeline proceeds to
   build the OCI image. The build process uses a kaniko executor as a build engine.

7. **Pushing the OCI Image**: After the OCI image is built, it is then pushed to a specified OCI registry.

8. **Signing the OCI Image**: If the build was triggered by a push event in GitHub, the `oci-image-builder` pipeline uses the `signify`
   service to sign the OCI image.
   This step ensures the integrity and authenticity of the OCI image.

## GitHub OIDC Identity Token Claims

The OIDC token issued by GitHub's OIDC identity provider contains several claims that are crucial for the `oci-image-builder` pipeline.
These claims are used to identify the workflow triggering the build pipeline and to clone the appropriate version of the source code. This
is essential for SLC-29 compliance, as it ensures that the exact version of the code that was tested in the PR or for which the push was
merged is built.

The validity and integrity of the OIDC token must be validated in the `oci-image-builder` pipeline and fail the pipeline execution if validation fails.
Because the OIDC token uses the `JWT` format, it can be validated with a standard validation process against GitHub's OIDC identity provider.

### Workflow Identification Claims

The OIDC token contains the following claims that can be used to identify the workflow that triggered the build pipeline. These include:

<!-- markdown-link-check-disable -->

- **iss**: The issuer of the token. This is always https://token.actions.githubusercontent.com. <!-- markdown-link-check-enable-->
- **iat**: The time when the token is issued.
- **exp**: The time when the token expires.
- **jti**: A unique identifier for the token.
- **nbf**: The time before which the token must not be accepted for processing.
- **kid**: The key ID of the key used to sign the token.
- **alg**: The algorithm used to sign the token.
- **run_id**: The ID of the workflow run.
- **run_number**: The number of the workflow run.
- **actor**: The login of the user who initiates the workflow run.
- **event_name**: The name of the event that triggers the workflow run.
- **workflow**: The name of the workflow that triggers the workflow run.
- **workflow_ref**: The git ref associated with the workflow file.
- **repository**: The repository where the workflow run occurs.
- **repository_owner**: The owner of the repository where the workflow run occurs.

### Source Code Cloning Claims

The OIDC token also contains claims that can be used to clone the appropriate version of the source code:

- **repository**: The repository where the workflow run occurs.
- **repository_owner**: The owner of the repository where the workflow run occurs.
- **event_name**: The name of the event that triggers the workflow run.
- **ref**: The git ref associated with the workflow run. This can be used to checkout the correct branch, tag, or commit.
- **ref_type**: The type of git ref associated with the workflow run. This can be used to determine whether the ref is a branch, tag, or
  commit.
- **base_ref**: The base git ref associated with the workflow run. This can be used to determine the base branch for a pull request.
- **head_ref**: The head git ref associated with the workflow run. This can be used to determine the head branch for a pull request.

These claims ensure that the `oci-image-builder` pipeline builds the exact version of the code that was provided in the PR or merged to the
branch, adhering to SLC-29 compliance.

## Parameters for the `oci-image-builder` Pipeline

The `oci-image-builder` pipeline requires certain data to be provided in parameters to execute a build process.
Certain parameters need to be defined by the user in addition to the data taken from the OIDC token.

These parameters include user-defined parameters, parameters from the OIDC token, and computed parameters.

### User Defined Parameters

- **Context**: The context of the build.
- **Dockerfile**: The Dockerfile to be used for the build.
- **Name**: The name of the image.
- **BuildArgs**: The build arguments to be passed to the build.
- **Tags**: The tags to be applied to the image.
- **ExportTags**: Whether to export the tags.

### Parameters from OIDC Token

- **RepoName**: The name of the repository.
- **RepoOwner**: The owner of the repository. Possible values include `kyma-project` and `kyma-incubator`.
- **JobType**: The type of job. Possible values include `presubmit` and `postsubmit`.
- **PullBaseSHA**: The base SHA of the pull request.
- **PullPullSHA**: The SHA of the pull request.

### Computed Parameters

The data for this parameter is not available in the OIDC token and should not be provided by the user.
We can extract it from the `ref` claim of the OIDC token.

- **PullNumber**: The number of the pull request.

## ProwJob Validation in Image Builder

In the scenario where the `oci-image-builder` pipeline is triggered from a ProwJob instead of a GitHub workflow, the validation process
differs slightly due to the absence of an OIDC token. Triggering the `oci-image-builder` pipeline from a ProwJob is considered a deprecated
scenario and must be supported only for backward compatibility and migration purposes.

### Process Flow

1. **Trigger ProwJob**: The user or automation triggers a ProwJob.

2. **Trigger oci-image-builder pipeline**: The ProwJob running Image builder client calls ADO API to trigger the `oci-image-builder` pipeline.
   The parameters required by the `oci-image-builder` pipeline are passed as parameters to the pipeline. These parameters' values are
   collected from data defined by the user and Prow environment variables. Together with these parameters, the parameters' hash signed by Prow is
   also passed to the pipeline as a parameter.

3. **Validating the Parameters**: The `oci-image-builder` pipeline, running in Azure DevOps, detects the call originating from Prow and
   validates the parameters against a hash signed by Prow. This step ensures that the parameters are valid and have not been tampered with.
   The absence of an OIDC token is used to detect that the call is coming from a ProwJob.

### Parameters Validation

The parameters passed to the `oci-image-builder` pipeline are validated against a hash signed by Prow. This hash is computed based on the
names and values of the parameters. The hash value is signed with a secret known to the `Prow` and the `oci-image-buidler` pipeline only.
The `oci-image-builder` pipeline verifies this hash to ensure that the parameters have not been tampered with and are indeed coming from the
ProwJob.

This validation process ensures the secure and authorized passing of information about the ProwJob and the image to build. It provides a
reliable infrastructure for the building of OCI images when the pipeline is triggered from a ProwJob.

## Block Diagram

![image-builder-block-diagram](documentation_assets/image-builder-block-diagram.png)

## Activity Diagram

![image-builder-activity-diagram](documentation_assets/image-builder-activity-diagram.png)

## Conclusion

The Image Builder solution, with its seamless integration with GitHub workflows and Azure DevOps pipeline, offers developers a robust and
secure method to incorporate the building of OCI images into their workflows. By leveraging a signed JWT format in which an OIDC token from
GitHub's OIDC identity provider is passed, it ensures the secure and authorized passing of information about the workflow and the image to
build. The entire build process adheres to SLC-29 compliance, providing a reliable infrastructure for the building of OCI images.
