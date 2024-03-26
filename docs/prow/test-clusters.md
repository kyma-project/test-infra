# Prow Test Clusters

This document gathers information about test clusters that Prow jobs build. All test clusters are built in the `sap-kyma-prow-workloads` project.


## Cluster Authorization

### Permissions

Kyma developers are gathered in the `kyma_developers@sap.com` Google Group administrated by the [Neighbors team](https://github.com/orgs/kyma-project/teams/prow/members?utf8=%E2%9C%93&query=role%3Amaintainer). All the group permissions are managed in IAM, where the `kyma_developers@sap.com` group has the **kyma_developer** role assigned so that its members can access test clusters and VMs in read-only mode.

### Custom Permissions

Previously, you could only grant roles to Google Cloud user accounts or Cloud IAM service accounts. Google Groups for GKE (Beta) now allows you to grant roles to the members of a G Suite Google Group. With this mechanism, the users and groups themselves are maintained by your G Suite administrators, completely outside of Kubernetes or Cloud Console.

Google Groups give you the possibility to gather Kyma Developers accounts and manage GCP Project permissions based on the group name. Additionally, you can grant Kubernetes Roles, Cluster Roles, Role Bindings, and Cluster Role Bindings to the specific Google Group on your cluster.

For example, all members of the `kyma_developers@sap.com` group receive the **cluster-admin** Cluster Role on the Kyma release cluster built by the **post-relXX-kyma-release-candidate** Prow job.

If you want to leverage this solution, [raise an issue](https://github.com/kyma-project/test-infra/issues/new/choose) with the Neighbors team. The process looks as follows:

1. The Neighbors team creates your custom Google Group, such as `your_custom_group@sap.com`, and adds it as a member of `gke-security-groups@sap.com`.

    ![dashboards](/docs/prow/assets/GGroups.png)

2. You write a test pipeline where you build the cluster with an additional parameter called **--security-group="gke-security-groups@sap.com**. 

3. In the next step of your test pipeline you create Cluster Role Bindings for the `your_custom_group@sap.com` custom group:

    ```
    kubectl create clusterrolebinding kyma-developers-group-binding --clusterrole="cluster-admin" --group="your_custom_group@sap.com"
    ```

When you complete all the steps, members of your custom group are able to access the cluster with elevated privileges.
