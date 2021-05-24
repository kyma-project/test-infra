import base64
import json
import os

from slack_bolt import App
from slack_sdk.errors import SlackApiError


def main(event, context):
	# Using SLACK_BOT_TOKEN environment variable
	app = App(
	)
	# Set Slack API base URL to the URL of slack-connector application gateway.
	app.client.base_url = "{}/".format(
		os.environ['OOM_FOUND_SLACK_CONNECTOR_{}_GATEWAY_URL'.format(os.environ['SLACK_API_ID']).replace('-', '_')]
	)
	print("received message with id: {}".format(event["data"]["MessageId"]))
	print("Slack api base URL: {}".format(app.client.base_url))
	print("sending notification to channel: {}".format(os.environ['NOTIFICATION_SLACK_CHANNEL']))
	# Get cloud events data.
	msg = json.loads(base64.b64decode(event["data"]["Data"]))
	print("msg: {}".format(msg))
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
														"text": "OutOfMemory event found in <{}|{}> prowjob.".format(
															msg["url"], msg["job_name"])
													}
												}
											])
		assert result["ok"]
		print("sent notification for message with id: {}".format(event["data"]["MessageId"]))
	except SlackApiError as e:
		assert result["ok"] is False
		print(f"Got an error: {e.response['error']}")
		print("failed sent notification for message with id: {}".format(event["data"]["MessageId"]))
