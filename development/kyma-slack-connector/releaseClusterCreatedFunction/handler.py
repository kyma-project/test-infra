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
	slack_channel = os.environ['NOTIFICATION_SLACK_CHANNEL']
	env_prefix = os.environ['ENV_PREFIX']
	base_url = os.environ['{}_SLACK_CONNECTOR_{}_GATEWAY_URL'.format(env_prefix, slack_api_id)]
	# Set Slack API base URL to the URL of slack-connector application gateway.
	app.client.base_url = "{}/".format(base_url)
	print("received message with id: {}".format(event["data"]["ID"]))
	print("Slack api base URL: {}".format(app.client.base_url))
	print("sending notification to channel: {}".format(slack_channel))
	# Get cloud events data.
	msg = json.loads(base64.b64decode(event["data"]["Data"]))
	uploadedKubeconfig=[]
	result=[]
	
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
		kubeconfigFilename = "kubeconfig-"+msg["cluster_name"]+".yaml"
		uploadedKubeconfig = app.client.files_upload(content=msg["kubeconfig"],
														filename=kubeconfigFilename,
														channels=slack_channel,
														thread_ts=result["message"]["ts"],
														initial_comment="Kubeconfig for the `{}` cluster: :blobwant:".format(
	 																	msg["kyma_version"],
	 																	msg["cluster_name"])
		)
		assert uploadedKubeconfig["ok"]
		print("uploaded kubeconfig for message id: {}".format(event["data"]["ID"]))
	except SlackApiError as e:
		assert uploadedKubeconfig["ok"] is False
		print(f"Got an error: {e.response['error']}")
		print("failed upload file for message id: {}".format(event["data"]["ID"]))
		return
