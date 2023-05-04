import os

import yaml
from github import Github
from slack_bolt import App
from slack_sdk.errors import SlackApiError

TOOLS_GITHUB_TEST_INFRA_REPO = "kyma/test-infra"
TOOLS_GITHUB_HOST = "github.tools.sap"
USERS_MAP_FILE_PATH = "users-map.yaml"
USERS_MAP_FILE_REF = "main"


def get_slack_username(github_username: str, users_map: list) -> str:
    for item in users_map:
        try:
            if github_username == item["sap.tools.github.username"]:
                return item["com.slack.enterprise.sap.username"]
        except KeyError:
            print(f"KeyError: sap.tools.github.username or com.slack.enterprise.sap.username not found in {item}")
    return None


def main(event, context):
    slack_bot_token = os.environ['SLACK_BOT_TOKEN']
    slack_channel = os.environ['NOTIFICATION_SLACK_CHANNEL']
    tools_github_bot_token = os.environ['TOOLS_GITHUB_BOT_TOKEN']
    app = App(token=slack_bot_token)

    # Create Github Enterprise client with custom hostname to access the users-map.yaml file.
    ghclient = Github(base_url=f"https://{TOOLS_GITHUB_HOST}/api/v3", login_or_token=tools_github_bot_token)
    repo = ghclient.get_repo(TOOLS_GITHUB_TEST_INFRA_REPO)
    content = repo.get_contents(USERS_MAP_FILE_PATH, ref=USERS_MAP_FILE_REF)

    # Read users-map.yaml file content.
    users_map = yaml.load(content.decoded_content.decode(), Loader=yaml.FullLoader)
    # Find the sender sap.tools.github.username in the users_map list and return the com.slack.enterprise.sap.username
    sender_slack_username = get_slack_username(event["data"]["sender"]["login"], users_map)
    # Find the assignee sap.tools.github.username in the users_map list and return the com.slack.enterprise.sap.username
    # If the assignee is not set, return None
    try:
        assignee_slack_username = get_slack_username(event["data"]["issue"]["assignee"]["login"], users_map)
    except TypeError:
        assignee_slack_username = None

    print("Received event of type issuesevent.labeled")

    # Parse the event data to get details for constructing notification message.
    msg = event["data"]
    label = msg["label"]["name"]
    title = msg["issue"]["title"]
    number = msg["issue"]["number"]
    repo = msg["repository"]["name"]
    org = msg["repository"]["owner"]["login"]
    if assignee_slack_username:
        assignee = f"Issue #{number} in repository {org}/{repo} is assigned to <@{assignee_slack_username}>"
    else:
        assignee = f"Issue #{number} in repository {org}/{repo} is not assigned."
    if sender_slack_username:
        sender = f"<@{sender_slack_username}>"
    else:
        sender = msg["sender"]["login"]
    issue_url = msg["issue"]["html_url"]

    # Run only for internal-incident and customer-incident labels
    if (label == "internal-incident") or (label == "customer-incident"):
        print(f"Label matched, Sending notifications to channel: {slack_channel}")
        try:
            # Build and deliver message to the channel.
            result = app.client.chat_postMessage(channel=slack_channel,
                                                 text=f"issue {title} #{number} labeld as {label} in {repo}",
                                                 username="GithubBot",
                                                 blocks=[
                                                     {
                                                         "type": "context",
                                                         "elements":
                                                             [
                                                                 {
                                                                     "type": "image",
                                                                     "image_url": "https://mpng.subpng.com/20180802/bfy/kisspng-portable-network-graphics-computer-icons-clip-art-caribbean-blue-tag-icon-free-caribbean-blue-pric-5b63afe8224040.3966331515332597521403.jpg",
                                                                     "alt_text": "label"
                                                                 },
                                                                 {
                                                                     "type": "mrkdwn",
                                                                     "text": "SAP Github issue labeled"
                                                                 }
                                                             ]
                                                     },
                                                     {
                                                         "type": "header",
                                                         "text": {
                                                             "type": "plain_text",
                                                             "text": f"SAP Github {label}"
                                                         }
                                                     },
                                                     {
                                                         "type": "section",
                                                         "text":
                                                             {
                                                                 "type": "mrkdwn",
                                                                 "text": f"@here {sender} labeled issue `{title}` as `{label}`.\n{assignee} <{issue_url}|See issue here.>"
                                                             }
                                                     },
                                                 ])
            assert result["ok"]
            print(f"Sent notification for issue #{number}")
        except SlackApiError as e:
            print(f"Got an error: {e.response['error']}")
            print(f"Failed sent notification for issue #{number}")
    else:
        print(f"Label {label} is not supported, ignoring.")
