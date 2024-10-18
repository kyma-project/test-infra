'''This function can receive various data types and sends Slack messages'''

# common-slack-bot-token
# google logging https://cloud.google.com/run/docs/logging#writing_structured_logs
# python wsgi pep https://peps.python.org/pep-3333/#environ-variables
# gunicorn https://docs.gunicorn.org/en/stable/run.html#
# flask app: https://flask.palletsprojects.com/en/2.2.x/quickstart/#a-minimal-application
# flask request docs: https://flask.palletsprojects.com/en/2.2.x/api/#incoming-request-data
# flask responses docs: https://flask.palletsprojects.com/en/2.2.x/quickstart/#about-responses
# slack app send message: https://api.slack.com/messaging/sending

import json
import os
import sys
import traceback
import base64
from typing import Dict, Any
from flask import Flask, request, make_response, Response
from cloudevents.http import from_http  # type: ignore
from slack_bolt import App
from slack_sdk import WebClient
from slack_sdk.errors import SlackApiError


class LogEntry(dict):
    '''LogEntry simplifies logging by returning JSON string'''

    def __str__(self):
        return json.dumps(self)


app = Flask(__name__)
project_id: str = os.getenv('PROJECT_ID', '')
component_name: str = os.getenv('COMPONENT_NAME', '')
application_name: str = os.getenv('APPLICATION_NAME', '')
slack_channel_id: str = os.getenv('PROW_DEV_NULL_SLACK_CHANNEL_ID', '')
slack_release_channel_id: str = os.getenv('RELEASE_SLACK_CHANNEL_ID', '')
slack_team_channel_id: str = os.getenv('KYMA_TEAM_SLACK_CHANNEL_ID', '')
slack_base_url: str = os.getenv('SLACK_BASE_URL', '')  # https://slack.com/api
kyma_security_slack_group_name: str = os.getenv('KYMA_SECURITY_SLACK_GROUP_NAME', '')
# TODO: make it configurable through env vars
with open('/etc/slack-secret/common-slack-bot-token-test', encoding='utf-8') as token_file:
    slack_bot_token = token_file.readline()
slack_app = App(token=slack_bot_token)
slack_client = WebClient(token=slack_bot_token)

slack_usergroups = slack_app.client.usergroups_list()
tmp_groups = [usergroup["id"] for usergroup in slack_usergroups["usergroups"] if
              usergroup["handle"] == "btp-kyma-security"]
if len(tmp_groups) != 1:
    print(LogEntry(
        severity="ERROR",
        message=(
            "Failed get kyma security slack group id from usersgroups, "
            f"got unexpected number of items, expected 1 but got {len(tmp_groups)}"
        )
    ))
kyma_security_slack_group_id: str = tmp_groups[0]


def prepare_log_fields() -> Dict[str, Any]:
    '''prepare_log_fields prapares basic log fields'''
    log_fields: Dict[str, Any] = {}
    request_is_defined = "request" in globals() or "request" in locals()
    if request_is_defined and request:
        trace_header = request.headers.get("X-Cloud-Trace-Context")
        if trace_header and project_id:
            trace = trace_header.split("/")
            log_fields["logging.googleapis.com/trace"] = f"projects/{project_id}/traces/{trace[0]}"
    log_fields["Component"] = "slack-message-sender"
    log_fields["labels"] = {
        "io.kyma.component": "slack-message-sender"
    }
    return log_fields


def get_pubsub_message():
    '''get_pubsub_message unpacks pubsub message and does basic type checks'''
    envelope = request.get_json()
    if not envelope:
        # pylint: disable=broad-exception-raised
        raise Exception("no Pub/Sub message received")

    if not isinstance(envelope, dict) or "message" not in envelope:
        # pylint: disable=broad-exception-raised
        raise Exception("invalid Pub/Sub message format")

    pubsub_message = envelope["message"]
    return pubsub_message


def prepare_success_response() -> Response:
    '''prepare_success_response return success response'''
    resp = make_response()
    resp.content_type = 'application/json'
    resp.status_code = 200
    return resp


def prepare_error_response(err: str, log_fields: Dict[str, Any]) -> Response:
    '''prepare_error_response return error response with stacktrace'''
    _, exc_value, _ = sys.exc_info()
    stacktrace = repr(traceback.format_exception(exc_value))
    print(LogEntry(
        severity="ERROR",
        message=f"Error: {err}\nStack:\n {stacktrace}",
        **log_fields,
    ))
    resp = make_response()
    resp.content_type = 'application/json'
    resp.status_code = 500
    return resp


@app.route("/secret-leak-found", methods=["POST"])
def secret_leak_found() -> Response:
    '''secret_leak_found handles found secret leak Slack messages'''
    log_fields: Dict[str, Any] = prepare_log_fields()
    log_fields["labels"]["io.kyma.app"] = "secret-leak-found"

    # create a CloudEvent
    event = from_http(request.headers, request.get_data())
    print(LogEntry(
        severity="DEBUG",
        message=f"event data: {event.data}",
        **log_fields,
    ))

    try:
        print(LogEntry(
            severity="INFO",
            message=f"Sending notification to {slack_channel_id}.",
            **log_fields,
        ))

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
        print(LogEntry(
            severity="INFO",
            message=f'Slack message send, message id: {result["ts"]}',
            **log_fields,
        ))
        result = slack_app.client.chat_postMessage(
            channel=slack_channel_id,
            text=f"<!subteam^{kyma_security_slack_group_id}>, just to let you know we got this.\n",
            username="KymaBot",
            thread_ts=result["ts"],
            unfurl_links=True,
            unfurl_media=True,
            link_names=True,
            blocks=[{
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": f"<!subteam^{kyma_security_slack_group_id}>, just to let you know we got this.\n"
                }
            }]
        )
        print(LogEntry(
            severity="INFO",
            message=f'Slack message send, message id: {result["ts"]}',
            **log_fields,
        ))
        return prepare_success_response()
    # pylint: disable=broad-exception-caught
    except Exception as err:
        return prepare_error_response(str(err), log_fields)


@app.route("/release-cluster-created", methods=["POST"])
def release_cluster_created() -> Response:
    '''this function sends kubeconfig in a Slack channel for newly created release cluster'''
    log_fields: Dict[str, Any] = prepare_log_fields()
    log_fields["labels"]["io.kyma.app"] = "release-cluster-created"
    try:
        pubsub_message = get_pubsub_message()
        if isinstance(pubsub_message, dict) and "data" in pubsub_message:
            release_info = json.loads(base64.b64decode(pubsub_message["data"]).decode("utf-8").strip())
            print(LogEntry(
                severity="INFO",
                message=f"Sending notification to {slack_release_channel_id}.",
                **log_fields,
            ))

            result = slack_app.client.chat_postMessage(
                channel=slack_release_channel_id,
                text=f"Kyma {release_info['kyma_version']} was released.",
                username="ReleaseBot",
                unfurl_links=True,
                unfurl_media=True,
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
                            "text": f"Kyma OS {release_info['kyma_version']} was released :tada:"
                        }
                    },
                    {
                        "type": "section",
                        "text": {
                            "type": "mrkdwn",
                            "text": f"Kubeconfig for the `{release_info['cluster_name']}` cluster is in the thread"
                        }
                    }
                ],
            )
            print(LogEntry(
                severity="INFO",
                message=f'Slack message send, message id: {result["ts"]}',
                **log_fields,
            ))

            kubeconfig_filename = f"kubeconfig-{release_info['cluster_name']}.yaml"
            uploaded_kubeconfig = slack_app.client.files_upload(
                content=release_info["kubeconfig"],
                filename=kubeconfig_filename,
                channels=slack_release_channel_id,
                thread_ts=result["message"]["ts"],
                initial_comment=f"Kubeconfig for the `{release_info['cluster_name']}` cluster: :blobwant:"
            )
            print(LogEntry(
                severity="INFO",
                message=f'Slack message send, message id: {uploaded_kubeconfig["ts"]}',
                **log_fields,
            ))

            return prepare_success_response()

        return prepare_error_response("Cannot parse pubsub data", log_fields)
    # pylint: disable=broad-exception-caught
    except Exception as err:
        return prepare_error_response(str(err), log_fields)


def get_slack_user_mapping():
    '''Fetches Slack users and returns a mapping of real names to Slack IDs'''
    try:
        users = {}
        response = slack_client.users_list()
        for member in response['members']:
            real_name = member.get('real_name')
            slack_id = member.get('id')
            if real_name and slack_id:
                users[real_name] = slack_id
        return users
    except SlackApiError as e:
        print(LogEntry(
            severity="ERROR",
            message=f"Error fetching Slack users: {e.response['error']}",
            **prepare_log_fields(),
        ))
        return {}


# Cache the user mapping to avoid frequent API calls
slack_user_mapping = get_slack_user_mapping()


@app.route("/issue-labeled", methods=["POST"])
def issue_labeled() -> Response:
    '''This function sends information about labeled issues in a Slack channel'''
    log_fields: Dict[str, Any] = prepare_log_fields()
    log_fields["labels"]["io.kyma.app"] = "issue-labeled"
    try:
        pubsub_message = get_pubsub_message()
        if isinstance(pubsub_message, dict) and "data" in pubsub_message:
            payload = json.loads(base64.b64decode(pubsub_message["data"]).decode("utf-8").strip())

            label = payload["label"]["name"]
            if label in ("internal-incident", "customer-incident"):
                title = payload["issue"]["title"]
                number = payload["issue"]["number"]
                repo = payload["repository"]["name"]
                org = payload["repository"]["owner"]["login"]
                issue_url = payload["issue"]["html_url"]

                assignee_name = payload.get("assigneeName")
                assignee_slack_id = None
                if assignee_name:
                    assignee_slack_id = slack_user_mapping.get(assignee_name)

                if assignee_slack_id:
                    assignee_text = f"Issue #{number} in repository {org}/{repo} is assigned to <@{assignee_slack_id}>."
                else:
                    assignee_text = f"Issue #{number} in repository {org}/{repo} is not assigned."

                sender_name = payload.get("senderName")
                sender_slack_id = None
                if sender_name:
                    sender_slack_id = slack_user_mapping.get(sender_name)

                if sender_slack_id:
                    sender_text = f"<@{sender_slack_id}>"
                else:
                    sender_text = sender_name or "Someone"

                print(LogEntry(
                    severity="INFO",
                    message=f"Sending notification to {slack_team_channel_id}.",
                    **log_fields,
                ))

                result = slack_app.client.chat_postMessage(
                    channel=slack_team_channel_id,
                    text=f"Issue {title} #{number} labeled as {label} in {repo}",
                    username="GithubBot",
                    unfurl_links=True,
                    unfurl_media=True,
                    blocks=[
                        {
                            "type": "context",
                            "elements": [
                                {
                                    "type": "image",
                                    "image_url": "https://mpng.subpng.com/20180802/bfy/kisspng-portable-network-graphics-computer-icons-clip-art-caribbean-blue-tag-icon-free-caribbean-blue-pric-5b63afe8224040.3966331515332597521403.jpg",
                                    "alt_text": "label"
                                },
                                {
                                    "type": "mrkdwn",
                                    "text": "GitHub issue labeled"
                                }
                            ]
                        },
                        {
                            "type": "header",
                            "text": {
                                "type": "plain_text",
                                "text": f"GitHub {label}"
                            }
                        },
                        {
                            "type": "section",
                            "text": {
                                "type": "mrkdwn",
                                "text": f"@here {sender_text} labeled issue `{title}` as `{label}`.\n{assignee_text} <{issue_url}|See the issue here.>"
                            }
                        },
                    ],
                )
                print(LogEntry(
                    severity="INFO",
                    message=f'Slack message sent, message id: {result["ts"]}',
                    **log_fields,
                ))

            return prepare_success_response()

        return prepare_error_response("Cannot parse pubsub data", log_fields)
    except Exception as err:
        return prepare_error_response(str(err), log_fields)
