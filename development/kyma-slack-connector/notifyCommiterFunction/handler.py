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
			if commiter != "":
				print(commiter)
				if slack_users != "":
					slack_users = "{}, <@{}>".format(slack_users, commiter)
				else:
					slack_users = "<@{}>".format(commiter)
		if slack_users != "":
			notify_msg = "{} please check what's going on".format(slack_users)
		else:
			notify_msg = "<!here>, couldn't find commiter slack username, please check this failure or ask commiter for it."
	else:
		notify_msg = "<!here>, couldn't find commiter slack username, please check this failure or ask commiter for it."
	print(notify_msg)
	channel_name = "kyma-prow-dev-null"
	conversation_id = None
	thread_id = None
	try:
		# Call the conversations.list method using the WebClient
		for response in app.client.conversations_list():
			if conversation_id is not None:
				break
			for channel in response["channels"]:
				if channel["name"] == channel_name:
					conversation_id = channel["id"]
					#Print result
					print(f"Found conversation ID: {conversation_id}")
					break
# https://api.slack.com/messaging/retrieving
# https://api.slack.com/methods/conversations.history#response
# https://api.slack.com/methods/search.messages
# https://slack.dev/bolt-python/tutorial/getting-started
# https://api.slack.com/start/building/bolt-python
# https://api.slack.com/messaging/retrieving
# https://api.slack.com/methods/conversations.list
# https://console.cloud.google.com/logs/query;query=error_group%2528%22CPec6unB58T4nQE%22%2529%0AlogName:%22prowjobs%22%0Aresource.type%3D%22gce_instance%22%0Aresource.labels.instance_id%3D%225616924538911123302%22;timeRange=2021-10-01T13:57:51.670Z%2F2021-10-01T14:57:51.670Z?project=sap-kyma-prow&supportedpurview=project - this should not be an error in gcp.
	except SlackApiError as e:
		print(f"Error: {e}")
	conversation_history = []
	try:
	# Call the conversations.history method using the WebClient
	# conversations.history returns the first 100 messages by default
	# These results are paginated, see: https://api.slack.com/methods/conversations.history$pagination
		result = app.client.conversations_history(channel=conversation_id, limit=10)
		conversation_history = result["messages"]
	except SlackApiError as e:
		print("Error creating conversation: {}".format(e))
	for message in conversation_history:
		if msg["url"] in message["text"]:
			thread_id = message["ts"]
			break
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
