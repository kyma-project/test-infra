import os
from slack_sdk import WebClient
from slack_sdk.errors import SlackApiError


def main(event, context):
	print(os.environ['KYMA_SLACK_KYMA_SLACK_CONNECTOR_277A5551_00A9_49DB_9B9A_FBFD891BD070_GATEWAY_URL'])
	client = WebClient(base_url="{}/".format(os.environ['KYMA_SLACK_KYMA_SLACK_CONNECTOR_277A5551_00A9_49DB_9B9A_FBFD891BD070_GATEWAY_URL']))
	label = event["data"]["label"]["name"]
	print(label)
	title = event["data"]["issue"]["title"]
	print(title)
	number = event["data"]["issue"]["number"]
	print(number)
	repo = event["data"]["repository"]["name"]
	print(repo)
	try:
		assignee = "Issue {} in repository {} is assigned to `{}`.".format(number, repo, event["data"]["issue"]["assignee"]["login"])
	except TypeError:
		assignee = "Issue {} in repository {} is not assigned.".format(number, repo)
	print(assignee)
	sender = event["data"]["sender"]["login"]
	print(sender)
	issue_url = event["data"]["issue"]["html_url"]
	print(issue_url)
	# Run only for internal-incident and customer-incident labels
	if (label == "internal-incident") or (label == "customer-incident"):
		print("run postMessage")
		try:
			response = client.chat_postMessage(channel='kyma-prow-dev-null',
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
															"text": "*{}* labeled issue `{}` as `{}`.\n{} <{}|Check issue here.>".format(sender, title, label, assignee, issue_url)
														}
												},
												])
			#assert response["ok"]
			print(response)
		except SlackApiError as e:
			# You will get a SlackApiError if "ok" is False
			#assert e.response["ok"] is False
			#assert e.response["error"]  # str like 'invalid_auth', 'channel_not_found'
			print(f"Got an error: {e.response['error']}")
