import json
import os
import sys
import traceback
import base64
from typing import Dict, Optional, Any
from flask import Flask, request, make_response, Response
from cloudevents.http import from_http  # type: ignore
from slack_bolt import App
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


# Global user cache


user_cache: Dict[str, str] = {}


def get_user_id_by_username(username: str) -> Optional[str]:
    username_lower = username.lower()

    if username_lower in user_cache:
        return user_cache[username_lower]

    next_cursor = None

    while True:
        try:
            response = slack_app.client.users_list(limit=20, cursor=next_cursor)
        except SlackApiError as e:
            print(f"Slack API error: {e.response['error']}")
            return None

        users = response['members']

        for user in users:
            slack_username = user.get('name')
            real_name = user['profile'].get('real_name')
            user_id = user['id']

            if slack_username:
                user_cache[slack_username.lower()] = user_id
                if slack_username.lower() == username_lower:
                    return user_id

            if real_name:
                user_cache[real_name.lower()] = user_id
                if real_name.lower() == username_lower:
                    return user_id

        next_cursor = response.get('response_metadata', {}).get('next_cursor')
        if not next_cursor:
            break

    user_cache[username_lower] = None
    return None


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
            if label in ("internal-incident", "customer-incident", "neighbors-test"):
                title = payload["issue"]["title"]
                number = payload["issue"]["number"]
                repo = payload["repository"]["name"]
                org = payload["repository"]["owner"]["login"]
                issue_url = payload["issue"]["html_url"]

                # Find Slack ID for assignee
                assignee_login = payload["issue"].get("assignee", {}).get("login")
                if assignee_login:
                    assignee_slack_id = get_user_id_by_username(assignee_login)
                    if assignee_slack_id:
                        assignee_mention = f"<@{assignee_slack_id}>"
                    else:
                        assignee_mention = assignee_login  # Use GitHub login without mention
                else:
                    assignee_mention = "not assigned"

                # Find Slack ID for sender
                sender_login = payload["sender"]["login"]
                sender_slack_id = get_user_id_by_username(sender_login)
                if sender_slack_id:
                    sender_mention = f"<@{sender_slack_id}>"

                # Prepare assignee information
                assignee_info = f"Issue #{number} in repository {org}/{repo} is assigned to {assignee_mention}"

                # Send message to Slack
                print(LogEntry(
                    severity="INFO",
                    message=f"Sending notification to {slack_team_channel_id}.",
                    **log_fields,
                ))
                if not sender_slack_id:
                    result = slack_app.client.chat_postMessage(
                        channel=slack_team_channel_id,
                        text=f"Issue {title} #{number} labeled as {label} in {repo}",
                        username="GithubBot",
                        unfurl_links=True,
                        unfurl_media=True,
                        link_names=True,
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
                                        "text": "SAP GitHub issue labeled"
                                    }
                                ]
                            },
                            {
                                "type": "header",
                                "text": {
                                    "type": "plain_text",
                                    "text": f"SAP GitHub {label}"
                                }
                            },
                            {
                                "type": "section",
                                "text": {
                                    "type": "mrkdwn",
                                    "text": (
                                        f"{sender_mention} labeled issue `{title}` as `{label}`.\n"
                                        f"{assignee_info} <{issue_url}|See the issue here.>"
                                    )
                                }
                            },
                        ],
                    )
                    print(LogEntry(
                        severity="INFO",
                        message=f"Slack message sent, message id: {result['ts']}",
                        **log_fields,
                    ))

                    return prepare_success_response()

        return prepare_error_response("Cannot parse pubsub data", log_fields)
    except Exception as err:
        return prepare_error_response(str(err), log_fields)
