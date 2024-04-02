# Prow Runtime Images

This directory contains images that can be used as runtime images for all ProwJobs in Kyma's Prow Instance.

Refer to the `README.md` files in the image subdirectories for more information.

## Adding Additional Applications

To add additional applications into the images, open a pull request (PR) with changes. Follow these recommendations:
* Always build from a source to ensure compiler vulnerabilities do not affect the resulting binary.
* Link the binary to a specific version so that it's easier to update when necessary. 
* Build binaries in a separate stage, then copy the resulting binary into the final image to ensure images are small and contain the least number of layers.

## Write Image Tests

To write simple smoke tests with your image, add an executable file called `test.sh`.
The scripts should contain all steps that perform basic or advanced test operations against the image. You can use all binaries available in [E2E DinD K3d Image](./e2e-dind-k3d) to test the built image.
The test script must exit with a non-zero number if any steps have failed.

By default, current context of a test script is always Docker build context. Image name is passed as a variable `IMG`.

### Example

The example below showcases the example definition of the `test.sh` script.
```shell
#!/usr/bin/env bash
set -e
echo "$IMG"
docker run --rm $IMG -- some-command
test $? -eq 0 || exit 1
```