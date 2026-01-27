# Google Cloud Group Creation in the SAP Organization

This guide explains how to manage groups on the organization level on Google Cloud.
Our Terraform service accounts have project-level permissions, which are insufficient for managing groups at the organization level. This guide provides a manual workaround for creating and managing groups.

## Step-by-Step Guide

You can create a new group using either the Google Cloud Console or the `gcloud` command-line tool.

### Using The Google Cloud Console

1. Go to the [Google Cloud Console](https://console.cloud.google.com/).
2. Make sure you are in the `sap.com` organization.
3. In the navigation menu, go to **IAM & Admin** > **Groups**.
4. Click **Create Group**.
5. Fill in the group details:
    - **Name**: A descriptive name for the group.
    - **Group ID**: A unique ID for the group. This will be part of the group's email address.
    - **Description**: A brief description of the group's purpose.
6. Click **Create**.

### Using the `gcloud` Command-Line Tool

1. Create a group.
   ```bash
   gcloud identity groups create GROUP_ID@sap.com --organization=sap.com --display-name="GROUP_NAME" --description="GROUP_DESCRIPTION"
   ```
   - Replace `GROUP_ID` with the desired ID for your group.
   - Replace `GROUP_NAME` with the display name for your group.
   - Replace `GROUP_DESCRIPTION` with a description of your group.

## Configure Terraform Permissions

After creating the group, you must add the Terraform service accounts with the **Manager** role. This allows Terraform to manage the group's membership.

While this is a straightforward, one-step process in the Google Cloud Console, the `gcloud` CLI requires two separate commands. First, add the service account with the **Member** role, then modify the membership to grant the **Manager** role. This two-step approach is more reliable and helps prevent errors during the assignment process.

### Using the Google Cloud Console

1. Go to the group's page in the Google Cloud Console.
2. Click **Add Members**.
3. Add the following service accounts with the **Manager** role:
    - `terraform-executor@sap-kyma-prow.iam.gserviceaccount.com`
    - `terraform-planner@sap-kyma-prow.iam.gserviceaccount.com`
4. Click **Add**.

### Using the `gcloud` Command-Line Tool

1.  Add the service accounts with the **MEMBER** role.

   Replace `GROUP_EMAIL` with the email address of the group you created, and run the following command. The command adds the service accounts with the **MEMBER** role:

    ```bash
    gcloud identity groups memberships add \
      --group-email GROUP_EMAIL \
      --member-email terraform-executor@sap-kyma-prow.iam.gserviceaccount.com \
      --roles=MEMBER

    gcloud identity groups memberships add \
      --group-email GROUP_EMAIL \
      --member-email terraform-planner@sap-kyma-prow.iam.gserviceaccount.com \
      --roles=MEMBER
    ```

2.  To add the **MANAGER** role to the service accounts, replace `GROUP_EMAIL` with the email address of the group you created, and run the following command:

    ```bash
    gcloud identity groups memberships modify-membership-roles \
      --group-email GROUP_EMAIL \
      --member-email terraform-executor@sap-kyma-prow.iam.gserviceaccount.com \
      --add-roles=MANAGER

    gcloud identity groups memberships modify-membership-roles \
      --group-email GROUP_EMAIL \
      --member-email terraform-planner@sap-kyma-prow.iam.gserviceaccount.com \
      --add-roles=MANAGER
    ```
