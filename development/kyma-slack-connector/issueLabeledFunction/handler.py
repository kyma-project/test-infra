import os

from github import Github
from slack_bolt import App
from slack_sdk.errors import SlackApiError


def main(event, context):
	# Using SLACK_BOT_TOKEN environment variable
	slack_bot_token = os.environ['SLACK_BOT_TOKEN']
	slack_channel = os.environ['NOTIFICATION_SLACK_CHANNEL']
	app = App(token=slack_bot_token)
	# slack_api_id = os.environ['SLACK_API_ID'].replace('-', '_')
	# env_prefix = os.environ['ENV_PREFIX']
	# base_url = os.environ['{}_SLACK_CONNECTOR_{}_GATEWAY_URL'.format(env_prefix, slack_api_id)]
	# Set Slack API base URL to the URL of slack-connector application gateway.
	# app.client.base_url = "{}/".format(base_url)

	# using an access token
	g = Github("access_token")

	# Github Enterprise with custom hostname
	g = Github(base_url="https://{hostname}/api/v3", login_or_token="access_token")

	print("Received event of type issuesevent.labeled")
	print("Using Slack api base URL: {}".format(app.client.base_url))
	msg = event["data"]
	label = msg["label"]["name"]
	title = msg["issue"]["title"]
	number = msg["issue"]["number"]
	repo = msg["repository"]["name"]
	org = msg["repository"]["owner"]["login"]
	try:
		assignee = "Issue #{} in repository {}/{} is assigned to `{}`.".format(number, org, repo, msg["issue"]["assignee"]["login"])
	except TypeError:
		assignee = "Issue #{} in repository {}/{} is not assigned.".format(number, org, repo)
	sender = msg["sender"]["login"]
	issue_url = msg["issue"]["html_url"]
	# Run only for internal-incident and customer-incident labels
	if (label == "internal-incident") or (label == "customer-incident"):
		print("Label matched, Sending notifications to channel: {}".format(os.environ['NOTIFICATION_SLACK_CHANNEL']))
		try:
			# Deliver message to the channel.
			result = app.client.chat_postMessage(channel=slack_channel,
											 text="issue {} #{} labeld as {} in {}".format(title, number, label, repo),
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
														"text": "SAP Github {}".format(label)
														}
												},
												{
													"type": "section",
													"text":
														{
															"type": "mrkdwn",
															"text": "@here {} labeled issue `{}` as `{}`.\n{} <{}|See issue here.>".format(sender, title, label, assignee, issue_url)
														}
												},
												])
			assert result["ok"]
			print("Sent notification for issue #{}".format(number))
		except SlackApiError as e:
			print(f"Got an error: {e.response['error']}")
			print("Failed sent notification for issue #{}".format(number))
	else:
		print(f"Label {label} is not supported, ignoring.")
