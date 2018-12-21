# Create Custom Image
The purpose of this document is to define how to create a new Google Compute Engine [custom image](https://cloud.google.com/compute/docs/images?authuser=1#custom_images) with required dependencies. The new image can be used to provision VM instances with all dependencies already installed.

> **NOTE:** To run the following script, please make sure that you are signed in to the Google Cloud project with administrative rights.

## Image creation process

The script will do the following:

1. Provision a new VM under VM instance in Google Compute Engine.
   The new VM instance will be named according to the pattern **kyma-deps-image-vm-<RANDOM_ID>**, and will be created in a random **zone** in Europe.

2. Move the [install-deps-debian.sh](./install-deps-debian.sh) script to the newly provisioned VM, and execute it.

3. Shut down the new VM after installing all the required dependencies specified in [install-deps-debian.sh](./install-deps-debian.sh).

4. Create a new **image** using `--source-disk` as **kyma-deps-image-vm-<RANDOM_ID>** under image family **custom-images** in Google Compute Engine. The new image will be named **kyma-deps-image-<Current_Date>**

5. Delete the provisioned VM **kyma-deps-image-vm-<RANDOM_ID>**.
