"""Module containing tests for signify secret rotator"""

import base64
import json
from typing import Any, Dict
import unittest
from unittest.mock import MagicMock, Mock, patch
from flask.testing import FlaskClient
import test_fixtures
from rotate_signify_certificate import app, get_pubsub_message, prepare_log_fields

project_id: str = "test-project"
component_name: str = "test_component"
application_name: str = "test_application"


class TestPubSubMessageHandler(unittest.TestCase):
    """Test cases for fucntion to handle pub sub message"""

    def setUp(self) -> None:
        self.app: FlaskClient = app.test_client()
        self.app.testing = True

    @patch("rotate_signify_certificate.project_id", project_id)
    @patch("rotate_signify_certificate.component_name", component_name)
    @patch("rotate_signify_certificate.application_name", application_name)
    def test_prepare_log_fields(self):
        """Test that prepare log fields based on current env variables and cloud trace context"""
        with app.test_request_context(
            "/", headers={"X-Cloud-Trace-Context": "trace-id/other-info"}
        ):
            log_fields: Dict[str, Any] = prepare_log_fields()
            self.assertIn("logging.googleapis.com/trace", log_fields)
            self.assertEqual(
                log_fields["logging.googleapis.com/trace"],
                f"projects/{project_id}/traces/trace-id",
            )
            self.assertEqual(log_fields["Component"], "signify-certificate-rotator")

    def test_get_pubsub_message_no_message(self):
        """It should raise exception when no pubsub message is received"""
        with app.test_request_context("/", json={}):
            with self.assertRaises(Exception):
                get_pubsub_message()

    def test_get_pubsub_message_invalid_format(self):
        """It should raise exception when invalid pubsub message is received"""
        with app.test_request_context("/", json={"invalid_key": "value"}):
            with self.assertRaises(Exception):
                get_pubsub_message()


class TestRotateSignifySecret(unittest.TestCase):
    """Tests for rotating signify secret"""

    @classmethod
    def setUpClass(cls) -> None:
        cls.app: FlaskClient = app.test_client()
        cls.app.testing = True

    def setUp(self):
        """Setup for each test"""
        self.pubsub_message = {
            "message": {
                "attributes": {"eventType": "SECRET_ROTATE"},
                "data": base64.b64encode(
                    json.dumps(
                        {"labels": {"type": "signify"}, "name": "test-secret"}
                    ).encode()
                ).decode(),
            }
        }

        self.secret_data = test_fixtures.mocked_secret_data
        self.mocked_token_response = test_fixtures.mocked_access_token_response
        self.mocked_cert_create_response = test_fixtures.mocked_cert_create_response

    @patch("requests.post")
    @patch("rotate_signify_certificate.set_secret")
    @patch("rotate_signify_certificate.get_secret")
    def test_rotate_signify_secret_success(
        self,
        mock_get_secret: Mock,
        mock_set_secret: Mock,
        mock_post: Mock,
    ):
        """Test success repsonse for valid pubsub"""
        mock_get_secret.return_value = self.secret_data
        mock_post.side_effect = [
            MagicMock(json=lambda: self.mocked_token_response),
            MagicMock(json=lambda: self.mocked_cert_create_response),
        ]

        response = self.app.post("/", json=self.pubsub_message)
        self.assertEqual(response.status_code, 200)
        mock_set_secret.assert_called_once()


if __name__ == "__main__":
    unittest.main()
