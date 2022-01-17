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
	print("Slack api base URL: {}".format(app.client.base_url))
	print("sending notification to channel: {}".format(os.environ['NOTIFICATION_SLACK_CHANNEL']))
	# Get cloud events data.
	msg = json.loads(base64.b64decode(event["data"]["Data"]))
	try:
		# Deliver message to the channel.
		result = app.client.chat_postMessage(channel=os.environ['NOTIFICATION_SLACK_CHANNEL'],
											text="oom found in <{}|{}> prowjob.".format(msg["url"], msg["job_name"]),
											username="ProwBot",
											icon_url="https://www.stickpng.com/img/download/580b57fbd9996e24bc43bdf6",
											blocks=[
												{
													"type": "context",
													"elements": [
														{
															"type": "mrkdwn",
															"text": "OutOfMemory event"
														}
													]
												},
												{
													"type": "header",
													"text": {
														"type": "plain_text",
														"text": "OOM event found"
													}
												},
												{
													"type": "section",
													"text": {
														"type": "mrkdwn",
														"text": "@here, OutOfMemory event found in <{}|{}> prowjob.".format(
															msg["url"], msg["job_name"])
													}
												}
											])
		assert result["ok"]
		print("sent notification for message id: {}".format(event["data"]["ID"]))
	except SlackApiError as e:
		assert result["ok"] is False
		print(f"Got an error: {e.response['error']}")
		print("failed sent notification for message id: {}".format(event["data"]["ID"]))
