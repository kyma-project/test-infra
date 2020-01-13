# Prow test clusters

This document gathers information about test clusters that are build by Prow jobs. All test clusters (a.k.a. workload clusters) are build in sap-kyma-prow-workloads project.


## Cluster authorization

### Permissions

Kyma developers are gathered in `kyma_developers@sap.com` Google Group administrated by [Neighbors team](https://github.com/orgs/kyma-project/teams/prow/members?utf8=%E2%9C%93&query=role%3Amaintainer). All the group permission are managed in IAM, where the `kyma_developers@sap.com` group has `kyma_developer` role assigned so that its members can access test clusters and VMs in read-only mode.

### Custom permissions

Previously, you could only grant roles to Google Cloud user accounts or Cloud IAM service accounts. Google Groups for GKE (Beta) now allows you to grant roles to the members of a G Suite Google Group. With this mechanism, the users and groups themselves are maintained by your G Suite administrators, completely outside of Kubernetes or Cloud Console.

Google Groups give you the possibility to gather Kyma Developers accounts and manage GCP Project permissions based on the group name. Additionally, you can grant Kubernetes Roles, ClusterRoles, RoleBindings, and ClusterRoleBindings to the specific Google Group on your cluster.

For example, all members of the `kyma_developers@sap.com` group receive the **cluster-admin** ClusterRole on the Kyma release cluster built by `post-relXX-kyma-release-candidate` Prow job.

If you want to leverage this solution, [raise issue](https://github.com/kyma-project/test-infra/issues/new/choose) with Neighbors team. The process used to look as follows:

1. Neighbors team creates your custom Google Group (eg. `your_custom_group@sap.com`) and adds it as a member of `gke-security-groups@sap.com`.

    ![dashboards](/docs/prow/assets/GGroups.png)

2. You write a test pipeline where you build the cluster with additional paremeter called [**--security-group="gke-security-groups@sap.com**](https://github.com/kyma-project/test-infra/blob/7b84900e56679fccfbc9e6839a85ade1dabe72bd/prow/scripts/cluster-integration/helpers/provision-gke-cluster.sh#L60). 

3. In the next step of your test pipeline you create ClusterRoleBindings for the `your_custom_group@sap.com` custom group:

    ```
    kubectl create clusterrolebinding kyma-developers-group-binding --clusterrole="cluster-admin" --group="your_custom_group@sap.com"
    ```

When all steps are completed members of your custom group are able to access the cluster with elevated privileges.