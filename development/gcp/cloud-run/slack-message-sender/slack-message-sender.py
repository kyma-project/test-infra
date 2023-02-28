# common-slack-bot-token
# google logging https://cloud.google.com/run/docs/logging#writing_structured_logs
# python wsgi pep https://peps.python.org/pep-3333/#environ-variables
# gunicorn https://docs.gunicorn.org/en/stable/run.html#
# flask app: https://flask.palletsprojects.com/en/2.2.x/quickstart/#a-minimal-application
# flask request docs: https://flask.palletsprojects.com/en/2.2.x/api/#incoming-request-data
# flask responses docs: https://flask.palletsprojects.com/en/2.2.x/quickstart/#about-responses
# slack app send message: https://api.slack.com/messaging/sending

from flask import Flask, request, make_response
from cloudevents.http import from_http
from slack_bolt import App
import json, os, sys, traceback
from typing import TypedDict, Generic


app = Flask(__name__)
project_id = os.getenv('PROJECT_ID')
component_name = os.getenv('COMPONENT_NAME')
application_name = os.getenv('APPLICATION_NAME')
slack_channel_id = os.getenv('SLACK_CHANNEL_ID')
slack_base_url = os.getenv('SLACK_BASE_URL')  # https://slack.com/api
kyma_security_slack_group_name = os.getenv('KYMA_SECURITY_SLACK_GROUP_NAME')
# TODO: make it configurable through env vars
with open('/etc/slack-secret/common-slack-bot-token') as token_file:
    slack_bot_token = token_file.readline()
slack_app = App(token=slack_bot_token)

slack_usergroups = slack_app.client.usergroups_list()
tmp_groups = [usergroup["id"] for usergroup in slack_usergroups["usergroups"] if usergroup["handle"] == "btp-kyma-security"]
if len(tmp_groups) != 1:
    entry = dict(
        severity="ERROR",
        message=f"Failed get kyma security slack gropup id from usersgroups, got unexpected number of items, " +
                f"expected 1 but got {len(tmp_groups)}"
    )
    print(json.dumps(entry))
kyma_security_slack_group_id: str = tmp_groups[0]


@app.route("/secret-leak-found", methods=["POST"])
def secret_leak_found():
    log_fields: TypedDict[str, str] = {}
    request_is_defined = "request" in globals() or "request" in locals()
    if request_is_defined and request:
        trace_header = request.headers.get("X-Cloud-Trace-Context")
        if trace_header and project_id:
            trace = trace_header.split("/")
            log_fields["logging.googleapis.com/trace"] = f"projects/{project_id}/traces/{trace[0]}"
    log_fields["Component"]: str = "MessageSender"
    log_fields["labels"]: TypedDict[str, str] = {
        "io.kyma.app": "KymaBotSlackApp",
        "io.kyma.component": "MessageSender"
    }

    # create a CloudEvent
    event = from_http(request.headers, request.get_data())
    entry = dict(
        severity="DEBUG",
        message=f"event data: {event.data}",
        **log_fields,
    )
    print(json.dumps(entry))
    try:
        entry = dict(
            severity="INFO",
            message=f"Sending notification to {slack_channel_id}.",
            **log_fields,
        )
        print(json.dumps(entry))

        result = slack_app.client.chat_postMessage(
            channel=slack_channel_id,
            text=f"Found secrets in {event.data['job_name']} {event.data['job_type']} prowjob logs.\n"
                 f"Please rotate secret and prevent further leaks.\n"
                 f"See details in Github issue {event.data['githubIssueURL']}.",
            username="KymaBot",
            # TODO: host icon on our infrastructure
            icon_url="https://assets.stickpng.com/images/580b57fbd9996e24bc43bdfe.png",
            unfurl_links=True,
            unfurl_media=True,
            link_names=True,
            blocks=[
                {
                    "type": "header",
                    "text": {
                        "type": "plain_text",
                        "text": "Secret leak found"
                    }
                },
                {
                    "type": "divider"
                },
                {
                    "type": "section",
                    "text": {
                        "type": "mrkdwn",
                        "text": f"Found secrets in {event.data['job_name']} {event.data['job_type']} prowjob logs.\n"
                                f"Please rotate secret and prevent further leaks.\n"
                                f"*See details in Github issue <{event.data['githubIssueURL']}|#{event.data['githubIssueNumber']}>.*"
                    },
                    "accessory": {
                        "type": "image",
                        "image_url": "https://assets.stickpng.com/images/5f42baae41b1ee000404b6f4.png",
                        "alt_text": "URGENT"
                    }
                },
                {
                    "type": "divider"
                }
            ]
        )
        entry = dict(
            severity="INFO",
            message=f'Slack message send, message id: {result["ts"]}',
            **log_fields,
        )
        print(json.dumps(entry))
        msg_ts = result["ts"]
        result = slack_app.client.chat_postMessage(
            channel=slack_channel_id,
            text=f"<!subteam^{kyma_security_slack_group_id}>, just to let you know we got this.\n",
            username="KymaBot",
            thread_ts=msg_ts,
            unfurl_links=True,
            unfurl_media=True,
            link_names=True,
            blocks=[{
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": f"<!subteam^{kyma_security_slack_group_id}>, just to let you know we got this.\n"
                    }
                }
            ]
        )
        entry = dict(
            severity="INFO",
            message=f'Slack message send, message id: {result["ts"]}',
            **log_fields,
        )
        print(json.dumps(entry))
        resp = make_response()
        resp.content_type = 'application/json'
        resp.status_code = 200
        return resp
    except Exception as e:
        exc_type, exc_value, exc_traceback = sys.exc_info()
        stacktrace = repr(traceback.format_exception(exc_value))
        entry = dict(
            severity="ERROR",
            message=f"Error: {e}\nStack:\n {stacktrace}",
            **log_fields,
        )
        print(json.dumps(entry))
        resp = make_response()
        resp.content_type = 'application/json'
        resp.status_code = 500
        return resp
