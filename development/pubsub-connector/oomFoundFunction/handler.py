import os
from slack_bolt import App
from google.cloud import firestore
from slack_sdk.errors import SlackApiError


def reportOOMevent(client: App.client, message: dict):
	try:
		result = client.chat_postMessage(channel=os.environ['OOM_SLACK_CHANNEL'],
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
													 "text": "OOM event found",
													 "emoji": "true"
												 }
											 },
											 {
												 "type": "section",
												 "text": {
													 "type": "mrkdwn",
													 "text": "OutOfMemory event found in {} prowjob".format(
														 message["job_name"])
												 }
											 }
										 ])
		assert result["ok"]
	except SlackApiError as e:
		assert result["ok"] is False
		print(f"Got an error: {e.response['error']}")


def getProwjobDoc(db: firestore.Client, prowjobid: str) -> dict:
	doc_ref = db.collection('prowjobs').document('{}'.format(prowjobid))
	doc = doc_ref.get()
	if doc.exists:
		return doc.to_dict()
	else:
		return None


#def getProwjobID():


def main(event, context):
	app = App(
	)
	app.client.base_url = "{}/".format(os.environ['OOM_FOUND_SLACK_CONNECTOR_2906b647_0DFE_4BF0_98E8_1C50D0348550_GATEWAY_URL'])
	# Project ID is determined by the GCLOUD_PROJECT environment variable
	# GOOGLE_APPLICATION_CREDENTIALS
	#db = firestore.Client()
	#getProwjobDoc(db, )
	reportOOMevent(app.client, event["data"])
