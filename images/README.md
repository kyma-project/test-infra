# Prow runtime images

This directory contains images that can be used as runtime images for all ProwJobs in Kyma's prow instance.

Please refer to the README.md files in the image subdirectories for more information.

## Adding additional applications

To add additional applications into the images, open a PR with changes. Keep in mind the following recommendations:
* Always build from source to ensure compiler vulnerabilities do not affect resulting binary
* Link binary to specific version, so it's easier to update it when needed
* Build binaries in separate stage then copy resulting binary into final image to ensure images are small and contain the least number of layers
