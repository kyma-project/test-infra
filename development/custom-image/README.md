# Create Custom Image
The purpose of this document is to define how to create a new Google Compute Engine [custom image](https://cloud.google.com/compute/docs/images?authuser=1#custom_images) with required dependencies. You can use the new image to provision virtual machine (VM) instances with all dependencies already installed.

> **NOTE:** To run the following script, make sure that you are signed in to the Google Cloud project with administrative rights.

## Image creation process

The script performs the following steps:

1. Provision a new VM under the VM instance in the Google Compute Engine.
   The new VM instance is named according to the **kyma-deps-image-vm-{RANDOM_ID}** pattern and and will be created in a random **zone** in Europe.

2. Move the [`install-deps-debian.sh`](./install-deps-debian.sh) script to the newly provisioned VM and execute it.

3. Shut down the new VM after installing all the required dependencies specified in [`install-deps-debian.sh`](./install-deps-debian.sh).

4. Create a new **image** using `--source-disk` as **kyma-deps-image-vm-{RANDOM_ID}** under the **custom-images** image family in the Google Compute Engine. This creates a new **kyma-deps-image-{Current_Date}** image.

5. Delete the provisioned **kyma-deps-image-vm-{RANDOM_ID}** VM.
