import os

import handler

# To test locally, set the environment variables and run this script. Environment variables to set:
# SLACK_BOT_TOKEN - Slack bot token with permissions to post messages
# to the channel specified in NOTIFICATION_SLACK_CHANNEL.
# TOOLS_GITHUB_BOT_TOKEN - GitHub bot token with permissions to read the users-map.yaml file
# from the kyma/test-infra repository.
# SENDER - Tools GitHub username of the sender of the issue labeled event.
# ASSIGNEE - Tools GitHub username of the assignee of the issue labeled event.
# NOTIFICATION_SLACK_CHANNEL - Slack channel ID to post the test message to.
if __name__ == "__main__":
    # Set NOTIFICATION_SLACK_CHANNEL environment variable
    sender = os.environ['SENDER']
    assignee = os.environ['ASSIGNEE']
    test_event = {
        "event-type": "issuesevent.labeled",
        "data": {
            "sender": {
                "login": sender
            },
            "issue": {
                "assignee": {
                    "login": assignee
                },
                "title": "Test issue",
                "number": 123,
                "html_url": "https://github.tools.sap/kyma/test-infra/issues/123",
            },
            "repository": {
                "name": "test-infra",
                "owner": {
                    "login": "kyma"
                }
            },
            "label": {
                "name": "internal-incident"
            }
        }
    }
    test_context = {}
    handler.main(test_event, test_context)
