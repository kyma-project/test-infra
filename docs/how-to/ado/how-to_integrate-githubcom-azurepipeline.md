# GitHub.com and Azure Pipeline (ADO) integration

## Motivation

We have to verify that the integration of an SAP Azure DevopsPipeline as a quality gate for github.com works.
It is expected that a simple Azure DevOps pipeline will be implemented and configured to become a quality gate for a PR of a public github.com repository. To do so:
- Create two simple Azure DevOps Pipelines, one that always fails, and another one that finishes successfully. 
- Create a public GitHub repository
- Configure these two pipelines as a quality gate for PRs to this repository (by using Webhooks etc.).
- Open a PR to the repository


## First steps

> **NOTE** It is not recommended to create a Hyperspace based managed pipeline as Hyperspace does not support github.com related git repository management.

We opened a feature [request](https://github.tools.sap/hyper-pipe/portal/issues/2303), but the responsible team declined it providing the following reason:  
*"We focus on internal development only. If GH.com would be easily feasible, we'd be happy to also support it. Today it is not in our focus and it also seems not to be so easy. We integrate many internal development tools, some of which are not enabled for GH.com. If GH.com is a strategic investment we would contribute our part as well, but since we integrate only existing solutions, we cannot be front-runners here."*

## Solution

A stable integration between github.com and Azure Pipeline (Azure DevOps) can be established in two ways:

</details>

<details>
<summary>Create a GitHub connection with GitHub PAT (personal access token)</summary>

This solution is similar to the Hyperspace-related solution on the GitHub Enterprise side. This means that after the integration, the GitHub side webhook URLs are configured with a system-generated unique identifier. The ID's name is **channelId**. It has a secret unknown to the user.  
Here is the common format of this URL with `dev.azure.com` secure URL prefix: **\<azure devops organisation\>/_apis/public/hooks/externalEvents?publisherId=github&channelId=\<system generated unique identifier\>&api-version=7.1-preview**.  
Here are GitHub PAT (personal access token) recommended scopes: repo, user, admin:repo_hook.
  ![pat scopes](./images/pat-recommended-scopes.png)

Pros:

- Stable connection
- Configured by the system
- No additional maintenance

Cons:

- **Unable to parse webhook payload (request body data) in Azure Pipeline**
- You need to manage the GitHub.com side PAT (expiration date - if any)
- You don't know the secret for channelId
</details>

</details>

<details>
<summary>Create Incoming WebHook connection on Azure DevOps side (call a general webhook from GitHub)</summary>

In this solution, you create a secure connection (Incoming WebHook) on the Azure DevOps side and generate a specific URL for this connection. This URL is used on the GitHub side in the WebHook section.  
Here is the common format of this url  with `dev.azure.com` secure URL prefix: **\<azure devops organisation\>/_apis/public/distributedtask/webhooks/\<incoming webhook name\>?api-version=7.1-preview**.

Pros:

- Stable connection
- **Able to parse webhook payload (request body data) in Azure Pipeline**
- No additional maintenance
- You don't need to manage the GitHub.com side PAT (expiration date - if any)

Cons:

- You need to do some manual steps to establish the connection between github.com and Azure Pipeline

For more information, see [the official documentation](https://learn.microsoft.com/en-us/azure/devops/pipelines/yaml-schema/?view=azure-pipelines).  

</details>  
<br> 

> **NOTE** The webhook payload (request body data) parsing is mandatory in this scenario. Therefore, the first option is not the right direction in this usecase.
In the second option, you can use Incoming WebHook connection on the Azure DevOps side, and you can call this specific URL from any webhook-ready system, such as github.com. This way, you can parse webhook payload (request body data) in runtime in Azure Pipeline. This method ensures executing conditional stages and steps of your pipeline. Thus it is described in the following section of this document.


## Configuration

### GitHub

1. Open [GitHub](https://github.com)
2. Log in to your account
3. To create a new repository, click **New**

   ![start-create-repo](./images/start-create-repo.png)

4. Enter your **Repository name**, and choose Public or Private; then click **Create repository**

   ![create-repo](./images/create-repo.png)

### Azure DevOps

5. Open Azure DevOps project on dev.azure.com

6. Navigate to **Project settings**

   ![project-settings](./images/project-settings.png)

7. Select **Service connections**

   ![service-connections](./images/service-connections.png)

8. Click **New service connection** in the top right-hand corner

   ![new-service-connection](./images/new-service-connection.png)

9. Type `incoming` in the search field to find **Incoming WebHook**, and click **Next**

   ![incomingwebhook](./images/incomingwebhook.png)

10. Provide the required data on the **New Incoming WebHook service connection** panel; click **Save** 

    - WebHook name
    - Secret (for instance, generated 64-character-long string with lowercase, uppercase, and numbers)
    - Service connection name
    - Check **Grant access permission to all pipelines** under Security

    ![incomindwebhook-data](./images/incomindwebhook-data.png)

### Local IDE

11. Your repository is created. Now, clone it to your computer. To do this, copy the repository URL

    ![clone-repo](./images/clone-repo.png)

12. Open your IDE, such as Visual Studio code and start cloning the repository (in VS Code: press F1, then find Git:Clone)

![git-clone](./images/git-clone.png)

13. Then paste the repository URL, and hit Enter

![git-clone-repo](./images/git-clone-repo.png)

14. Choose the destination folder for the repository

![repo-location](./images/repo-location.png)

15. When it is downloaded to your computer, you can open it in the current window or a new one

![open-in-ide](./images/open-in-ide.png)

16. In the project root directory, create the `azure-pipelines.yml` file. This is the soul of your pipeline

    ![new-file](./images/new-file.png)

17. Add the following simple content to `azure-pipelines.yml`. Customize it according to your `Incoming WebHook` connection parameter

Some explanation:
- trigger: **none** (you manage the execution of the pipeline by GitHub webhook calls)
- resources.webhooks.webhook: `Incoming WebHook`'s name
- resources.webhooks.connection: `Incoming WebHook`'s service connection name

> **CAUTION** Replace the `Incoming WebHook`'s name inside the stages and steps.

```yaml
# Using a general purpose pipeline for Azure
trigger: none

resources:
  webhooks:
    - webhook: githubcomwebhook
      connection: githubconnection

stages:
  - stage: PrerequisiteCheck
    jobs:
      - job: ActionValueCheck
        steps:
          - script: |
              echo 'Action:' ${{ convertToJson(parameters.githubcomwebhook.action) }}
              echo 'Is the Condition reopened, opened, ready_for_review, synchronize?:' ${{ in(parameters.githubcomwebhook.action, 'reopened','opened','ready_for_review','synchronize') }}
```

18. Save the file, and check the right branch before commiting and pushing

    ![check-branch](./images/check-branch.png)

19. Add to stage changes, commit, and push to remote

    ![push-changes](./images/push-changes.png)

### Azure DevOps

20. Go back to ADO and start to create the pipeline

    ![pipeline](./images/pipeline.png)

21. Click **New pipeline** in the top right-hand corner

    ![new-pipeline](./images/new-pipeline.png)

22. Select **GitHub** as source

    ![pipeline-source](./images/pipeline-source.png)

23. Connect to GitHub if necessary, then select **All repositories** and type your repository name in the filter field

    ![pipeline-gitrepository](./images/pipeline-gitrepository.png)

24. Select the repository, then choose the existing `yaml` file (`azure-pipelines.yml`) in the next step and continue

25. In the Review step, you can check the `yaml` file content before you create the pipeline

    ![pipeline-review](./images/pipeline-review.png)

26. Choose **Save** under the Run dropdown to save the pipeline

    ![pipeline-save](./images/pipeline-save.png)

27. The pipeline is ready for a run

    ![pipeline-saved](./images/pipeline-saved.png)

### GitHub

28.  To configure the GitHub side webhook, go to your repository and click **Settings** 

![repo-settings](./images/repo-settings.png)    

29.  If you have a pull_request related webhook, start to edit it. If you don't have a webhook, click **Add webhook** 

![repo-hook-add](./images/repo-hook-add.png)

30.  Provide the required data, then click **Add webhook / Update webhook** 

Some explanation:

- Payload URL: Configure the following URL according to your Azure DevOps related **Organisation** and **Incoming WebHook name** - **\<azure devops organisation\>/_apis/public/distributedtask/webhooks/\<incoming web hook name\>?api-version=7.1-preview**
- Content type: `application/json`
- Secret: Secret you used in **Incoming WebHook** connection
- Enable SSL verification
- Choose the events you want to trigger this webhook. Select `Let me select individual events`; scroll down to check `Pull requests`

    ![repo-hook-new01](./images/repo-hook-new01.png)
    ![repo-hook-new02](./images/repo-hook-new02.png)
    ![repo-hook-new03](./images/repo-hook-new03.png)

31.  Click **Recent Deliveries**, and check the ping request (which makes a health check)

![repo-hook-ping](./images/repo-hook-ping.png)

The pipeline configuration is complete.

### Local IDE

32. To check the pipeline, go back to your IDE, and create a new branch (for example, a feature branch). Make some code changes, for example, in the `README.md` file. Then push the changes to GitHub

33. Open your GitHub repository in your browser and start a pull request

    ![test-pullrequest](./images/test-pullrequest.png)

34. When the pull request is open (for example, waiting for the review), GitHub will start a webhook call. When the pull request is closed, there is another webhook call. Accordingly, there must be two webhook calls. 

- 34.1. GitHub side `Recent Deliveries`:

![test-pullrequestresult](./images/test-pullrequestresult.png)


- 34.2. Azure DevOps side `Runs` :

![test-pullrequest-ado](./images/test-pullrequest-ado.png)

 Check payload parsing:

![test-ado-opened](./images/test-ado-opened.png)
![test-ado-closed](./images/test-ado-closed.png)
