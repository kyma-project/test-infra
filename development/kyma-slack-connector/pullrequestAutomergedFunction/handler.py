import base64
import json
import os

from slack_bolt import App
from slack_sdk.errors import SlackApiError


def main(event, context):
    # Using SLACK_BOT_TOKEN environment variable
    app = App()
    slack_api_id = os.environ['SLACK_API_ID'].replace('-', '_')
    env_prefix = os.environ['ENV_PREFIX']
    base_url = os.environ['{}_SLACK_CONNECTOR_{}_GATEWAY_URL'.format(env_prefix, slack_api_id)]
    # Set Slack API base URL to the URL of slack-connector application gateway.
    app.client.base_url = "{}/".format(base_url)
    print("received message with id: {}".format(event["data"]["ID"]))
    print("using slack api base URL: {}".format(app.client.base_url))
    # Get cloud events data.
    msg = json.loads(base64.b64decode(event["data"]["Data"]))
    # Go through all target channels or users to send notification.
    for target in msg["ownersSlackIDs"]:
        # target is a channel where function will send notification.
        try:
            print(f"Sending notification to {target}.")
            result = app.client.chat_postMessage(channel=target,
                                                 text="PR: {} Repo: {}\{} Title: {} View PR: {}".format(
                                                     msg["prNumber"],
                                                     msg["prOrg"],
                                                     msg["prRepo"],
                                                     msg["prTitle"],
                                                     msg["prURL"]),
                                                 username="KymaBot",
                                                 link_names=True,
                                                 blocks=[
                                                     {
                                                         "type": "header",
                                                         "text": {
                                                             "type": "plain_text",
                                                             "text": "PR automatically merged"
                                                         }
                                                     },
                                                     {
                                                         "type": "section",
                                                         "text": {
                                                             "type": "mrkdwn",
                                                             "text": "*PR:* {}\n*Repo:* {}\{}\n*Title:* {}".format(
                                                                 msg["prNumber"],
                                                                 msg["prOrg"],
                                                                 msg["prRepo"],
                                                                 msg["prTitle"])
                                                         },
                                                         "accessory":
                                                             {
                                                                 "type": "button",
                                                                 "text": {
                                                                     "type": "plain_text",
                                                                     "text": "View PR"
                                                                 },
                                                                 "url": msg["prURL"],
                                                                 "style": "primary"
                                                             }
                                                     }
                                                 ])
            # Check we got OK response, otherwise fail.
            assert result.get("ok", False), "Assert response from slack API is OK failed. This is critical error."
            print("sent notification for incoming message id: {}".format(event["data"]["ID"]))
        # https://slack.dev/python-slack-sdk/api-docs/slack_sdk/errors/index.html#slack_sdk.errors.SlackApiError
        except SlackApiError as e:
            # https://slack.dev/python-slack-sdk/api-docs/slack_sdk/web/slack_response.html#slack_sdk.web.slack_response.SlackResponse
            # Check we got NOK slack api response.
            assert e.response.get("ok", False) is False, \
                "Assert response from slack API is not OK failed. This should not be error."
            print(f"Got an error: {e.response['error']}")
            print("failed sent notification for message id: {}".format(event["data"]["ID"]))
