import base64
import json
import os

from slack_bolt import App
from slack_sdk.errors import SlackApiError


def main(event, context):
	# Using SLACK_BOT_TOKEN environment variable
	app = App(
	)
	slack_api_id = os.environ['SLACK_API_ID'].replace('-', '_')
	env_prefix = os.environ['ENV_PREFIX']
	base_url = os.environ['{}_SLACK_CONNECTOR_{}_GATEWAY_URL'.format(env_prefix, slack_api_id)]
	# Set Slack API base URL to the URL of slack-connector application gateway.
	app.client.base_url = "{}/".format(base_url)
	print("received message with id: {}".format(event["data"]["ID"]))
	print("slack api base URL: {}".format(app.client.base_url))
	print("sending notification to channel: {}".format(os.environ['NOTIFICATION_SLACK_CHANNEL']))
	# Get cloud events data.
	msg = json.loads(base64.b64decode(event["data"]["Data"]))
	print(msg)
	if len(msg["slackCommitersLogins"]) > 0:
		slack_users = ""
		for commiter in msg["slackCommitersLogins"]:
			print(commiter)
			if slack_users != "":
				slack_users = "{}, <@{}>".format(slack_users, commiter)
			else:
				slack_users = "<@{}>".format(commiter)
		notify_msg = "{} please check what's going on".format(slack_users)
	else:
		notify_msg = "<!here>, couldn't find commiter slack username, please check this failure or ask commiter for it."
	print(notify_msg)
	try:
		# Deliver message to the channel.
		# https://slack.dev/python-slack-sdk/api-docs/slack_sdk/web/slack_response.html#slack_sdk.web.slack_response.SlackResponse
		result = app.client.chat_postMessage(channel=os.environ['NOTIFICATION_SLACK_CHANNEL'],
											text="{} prowjob {} execution failed, view logs: {}".format(msg["job_type"], msg["job_name"], msg["url"]),
											username="CiForceBot",
											icon_emoji="https://www.stickpng.com/img/download/580b57fbd9996e24bc43bdfe/image",
											link_names="true",
											blocks=[
												{
													"type": "header",
													"text": {
														"type": "plain_text",
														"text": "Prowjob execution failed"
													}
												},
												{
													"type": "section",
													"text": {
														"type": "mrkdwn",
														"text": "*Name:* {}\n*Type:* {}\n<{}|*View logs*>".format(
															msg["job_name"], msg["job_type"], msg["url"])
													}
												},
												{
													"type": "section",
													"text": {
														"type": "mrkdwn",
														"text": "{}".format(notify_msg)
													}
												}
											])
		assert result.get("ok", False), "Assert response from slack API is OK failed. This is critical error."
		print("sent notification for message id: {}".format(event["data"]["ID"]))
	# https://slack.dev/python-slack-sdk/api-docs/slack_sdk/errors/index.html#slack_sdk.errors.SlackApiError
	except SlackApiError as e:
		# https://slack.dev/python-slack-sdk/api-docs/slack_sdk/web/slack_response.html#slack_sdk.web.slack_response.SlackResponse
		assert e.response.get("ok", False) is False,\
			"Assert response from slack API is not OK failed. This should not be error."
		print(f"Got an error: {e.response['error']}")
		print("failed sent notification for message id: {}".format(event["data"]["ID"]))
