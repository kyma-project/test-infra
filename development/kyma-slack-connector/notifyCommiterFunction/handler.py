import base64
import json
import os

from slack_bolt import App
from slack_sdk.errors import SlackApiError


def main(event, context):
	# Using SLACK_BOT_TOKEN environment variable
	app = App()
	notify_msg = None
	slack_api_id = os.environ['SLACK_API_ID'].replace('-', '_')
	env_prefix = os.environ['ENV_PREFIX']
	base_url = os.environ['{}_SLACK_CONNECTOR_{}_GATEWAY_URL'.format(env_prefix, slack_api_id)]
	# Set Slack API base URL to the URL of slack-connector application gateway.
	app.client.base_url = "{}/".format(base_url)
	result = None
	print("received message with id: {}".format(event["data"]["ID"]))
	print("using slack api base URL: {}".format(app.client.base_url))
	print("using slack channel: {}".format(os.environ['NOTIFICATION_SLACK_CHANNEL']))
	# Get cloud events data.
	msg = json.loads(base64.b64decode(event["data"]["Data"]))
	# Uncomment to get printed received messages payload
	#print(msg)
	if "slackCommitersLogins" in msg:
		if len(msg["slackCommitersLogins"]) > 0:
			slack_users = ""
			for commiter in msg["slackCommitersLogins"]:
				if commiter != "":
					if slack_users != "":
						slack_users = "{}, <@{}>".format(slack_users, commiter)
					else:
						slack_users = "<@{}>".format(commiter)
			if slack_users != "":
				notify_msg = "{} please check what's wrong".format(slack_users)
			else:
				notify_msg = "<!here>, couldn't find commiter slack username, please check what's wrong or ask commiter for it."
		else:
			notify_msg = "<!here>, couldn't find commiter slack username, please check what's wrong or ask commiter for it."
	# Uncomment to get printed @notification message
	#print(notify_msg)
	# channel_name is a channel where function will search for messages to use threads.
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
					# Print result
					print(f"Found conversation ID: {conversation_id}")
					break
	except SlackApiError as e:
		print(f"Error: {e}")
	conversation_history = []
	try:
		# Call the conversations.history method using the WebClient
		# conversations.history returns the first 100 messages by default
		# These results are paginated, see: https://api.slack.com/methods/conversations.history$pagination
		response = app.client.conversations_history(channel=conversation_id, limit=10)
		conversation_history = response["messages"]
	except SlackApiError as e:
		print("Error creating conversation: {}".format(e))
	try:
		for message in conversation_history:
			if msg["url"] in message["text"]:
				if "thread_ts" in message:
					thread_id = message["thread_ts"]
				else:
					thread_id = message["ts"]
				print(f"Found matching message, sending notification in a thread, thread_id: {thread_id}.")
				result = app.client.chat_postMessage(channel=os.environ['NOTIFICATION_SLACK_CHANNEL'],
													 thread_ts=thread_id,
													 text="Created issue #{} https://github.com/kyma-test-infra-dev/kyma/issues/{}".format(
														 msg["githubIssueNumber"],
														 msg["githubIssueNumber"]),
													 username="CiForceBot",
													 link_names=True,
													 blocks=[
														 {
															 "type": "header",
															 "text": {
																 "type": "plain_text",
																 "text": "Github issue created"
															 }
														 },
														 {
															 "type": "section",
															 "text": {
																 "type": "mrkdwn",
																 # project and repostiory should be variabels provided by pubsub message.
																 "text": "<https://github.com/kyma-test-infra-dev/kyma/issues/{}|*Issue #{} created.*>".format(
																	 msg["githubIssueNumber"],
																	 msg["githubIssueNumber"])
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
				break
		if thread_id is None:
			print(f"Matching message not found, sending notification as a standalone message.")
			result = app.client.chat_postMessage(channel=os.environ['NOTIFICATION_SLACK_CHANNEL'],
												 text="{} prowjob {} execution failed, view logs: {}, issue #{}: https://github.com/kyma-test-infra-dev/kyma/issues/{}".format(
													 msg["job_type"],
													 msg["job_name"],
													 msg["url"],
													 msg["githubIssueNumber"],
													 msg["githubIssueNumber"]),
												 username="CiForceBot",
												 link_names=True,
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
															 "text": "*Name:* {}\n*Type:* {}\n<{}|*View logs*>\n<https://github.com/kyma-test-infra-dev/kyma/issues/{}|*Issue #{}*>".format(
																 msg["job_name"],
																 msg["job_type"],
																 msg["url"],
																 msg["githubIssueNumber"],
																 msg["githubIssueNumber"])
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
		print(f"sent notification for incoming message id: {event.data.ID}")
	# https://slack.dev/python-slack-sdk/api-docs/slack_sdk/errors/index.html#slack_sdk.errors.SlackApiError
	except SlackApiError as e:
		# https://slack.dev/python-slack-sdk/api-docs/slack_sdk/web/slack_response.html#slack_sdk.web.slack_response.SlackResponse
		assert e.response.get("ok", False) is False, \
			"Assert response from slack API is not OK failed. This should not be error."
		print(f"Got an error: {e.response['error']}")
		print("failed sent notification for message id: {}".format(event["data"]["ID"]))
