import os, base64, json
from slack_sdk import WebClient
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
	print("Slack api base URL: {}".format(app.client.base_url))
	# Get cloud events data.
	msg = json.loads(base64.b64decode(event["data"]["Data"]))
	print("msg: {}".format(msg))
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
		print("sending notification to channel: {}".format(os.environ['NOTIFICATION_SLACK_CHANNEL']))
		try:
			# Deliver message to the channel.
			result = app.client.chat_postMessage(channel=os.environ['NOTIFICATION_SLACK_CHANNEL'],
											 text="oom found in <{}|{}> prowjob.".format(msg["url"], msg["job_name"]),
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
			print("sent notification for message with id: {}".format(event["data"]["MessageId"]))
		except SlackApiError as e:
			assert result["ok"] is False
			print(f"Got an error: {e.response['error']}")
			print("failed sent notification for message with id: {}".format(event["data"]["MessageId"]))
