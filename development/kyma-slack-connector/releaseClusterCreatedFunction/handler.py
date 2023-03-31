import base64
import json
import os

from slack_bolt import App
from slack_sdk.errors import SlackApiError


def main(event, context):
	slack_bot_token = os.environ['SLACK_BOT_TOKEN']
	slack_channel = os.environ['NOTIFICATION_SLACK_CHANNEL']
	app = App(token=slack_bot_token)

	print("received message with id: {}".format(event["data"]["ID"]))
	print("sending notification to channel: {}".format(slack_channel))

	# Get cloud events data.
	msg = json.loads(base64.b64decode(event["data"]["Data"]))
	uploaded_kubeconfig = []
	result = []
	
	try:
		# Deliver message to the channel.
		result = app.client.chat_postMessage(channel=slack_channel,
											text="Kyma {} was released.".format(msg["kyma_version"]),
											username="ReleaseBot",
											unfurl_links="true",
											unfurl_media="true",
											blocks=[
												{
													"type": "context",
													"elements": [
														{
															"type": "mrkdwn",
															"text": "_Kyma OS was released_"
														}
													]
												},
												{
													"type": "header",
													"text": {
														"type": "plain_text",
														"text": "Kyma OS {} was released :tada:".format(
																msg["kyma_version"])
													}
												},
												{
													"type": "section",
													"text": {
														"type": "mrkdwn",
														"text": "Kubeconfig for the `{}` cluster is in the thread".format(
																msg["cluster_name"])
													}
												}
											],
		)
		assert result["ok"]
		print("sent notification for message id: {}".format(event["data"]["ID"]))
	except SlackApiError as e:
		assert result["ok"] is False
		print(f"Got an error: {e.response['error']}")
		print("failed sent notification for message id: {}".format(event["data"]["ID"]))

	try:
		# push kubeconfig
		kubeconfig_filename = "kubeconfig-"+msg["cluster_name"]+".yaml"
		uploaded_kubeconfig = app.client.files_upload(content=msg["kubeconfig"],
														filename=kubeconfig_filename,
														channels=slack_channel,
														thread_ts=result["message"]["ts"],
														initial_comment="Kubeconfig for the `{}` cluster: :blobwant:".format(
	 																	msg["kyma_version"],
	 																	msg["cluster_name"])
		)
		assert uploaded_kubeconfig["ok"]
		print("uploaded kubeconfig for cluster {} for message id: {}".format(msg["cluster_name"], event["data"]["ID"]))
	except SlackApiError as e:
		assert uploaded_kubeconfig["ok"] is False
		print(f"Got an error: {e.response['error']}")
		print("failed upload file for message id: {}".format(event["data"]["ID"]))
		return
