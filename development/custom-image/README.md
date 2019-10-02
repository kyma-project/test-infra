# Create Custom Image
The purpose of this document is to define how to create a new Google Compute Engine [custom image](https://cloud.google.com/compute/docs/images) with required dependencies. You can use the new image to provision virtual machine (VM) instances with all dependencies already installed.

> **NOTE:** To run the following script, make sure that you are signed in to the Google Cloud project with administrative rights.

## Image creation process

To run the script, use the `create-custom-image.sh` command. To set the image created by this script as a default custom image, add the `--default` flag to the command.

> **NOTE:** Adding the `--default` flag to this script adds the `default:yes` label to the created custom image. By default, the [`provision-vm-and-start-kyma.sh`](../../prow/scripts/provision-vm-and-start-kyma.sh) script selects the latest default custom image available in Kyma project to provision the VM instance.


The script performs the following steps:

1. Provision a new VM under the VM instance in the Google Compute Engine.
   The new VM instance is named according to the **kyma-deps-image-vm-{RANDOM_ID}** pattern and created in a random **zone** in Europe.

2. Move the [`install-deps-debian.sh`](./install-deps-debian.sh) script to the newly provisioned VM and execute it.

3. Shut down the new VM after installing all the required dependencies specified in [`install-deps-debian.sh`](./install-deps-debian.sh).

4. Create a new **image** using `--source-disk` as **kyma-deps-image-vm-{RANDOM_ID}** under the **custom-images** image family in the Google Compute Engine. This creates a new **kyma-deps-image-{Current_Date}** image.

5. Delete the provisioned **kyma-deps-image-vm-{RANDOM_ID}** VM.
