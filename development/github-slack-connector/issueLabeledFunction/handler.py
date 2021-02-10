import os
from slack_sdk import WebClient
from slack_sdk.errors import SlackApiError


def main(event, context):
	slack_channel = os.environ['NOTIFICATION_SLACK_CHANNEL']
	client = WebClient(base_url="{}/".format(os.environ['KYMA_SLACK_SLACK_CONNECTOR_85DED56E_303B_43B3_A950_8B1C3D519561_GATEWAY_URL']))
	label = event["data"]["label"]["name"]
	title = event["data"]["issue"]["title"]
	number = event["data"]["issue"]["number"]
	repo = event["data"]["repository"]["name"]
	org = event["data"]["repository"]["owner"]["login"]
	try:
		assignee = "Issue #{} in repository {}/{} is assigned to `{}`.".format(number, org, repo, event["data"]["issue"]["assignee"]["login"])
	except TypeError:
		assignee = "Issue #{} in repository {}/{} is not assigned.".format(number, org, repo)
	sender = event["data"]["sender"]["login"]
	issue_url = event["data"]["issue"]["html_url"]
	# Run only for internal-incident and customer-incident labels
	if (label == "internal-incident") or (label == "customer-incident"):
		print("sending message to {} channel".format(slack_channel))
		try:
			response = client.chat_postMessage(channel=slack_channel,
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
			assert response["ok"]
		except SlackApiError as e:
			# You will get a SlackApiError if "ok" is False
			assert e.response["ok"] is False
			print(f"Got an error: {e.response['error']}")
