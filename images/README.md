# Prow Runtime Images

This directory contains images that can be used as runtime images for all ProwJobs in Kyma's Prow Instance.

Refer to the `README.md` files in the image subdirectories for more information.

## Adding Additional Applications

To add additional applications into the images, open a pull request (PR) with changes. Follow these recommendations:
* Always build from a source to ensure compiler vulnerabilities do not affect the resulting binary.
* Link the binary to a specific version so that it's easier to update when necessary. 
* Build binaries in a separate stage, then copy the resulting binary into the final image to ensure images are small and contain the least number of layers.
# (2025-03-04)