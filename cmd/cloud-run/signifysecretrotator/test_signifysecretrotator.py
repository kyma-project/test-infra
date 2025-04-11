"""Package containing tests for the Signify Secret Rotator."""

import base64
import json
from typing import List
import unittest
from unittest.mock import patch


from flask import Flask
from cryptography import x509

from signifysecretrotator import (
    rotate_signify_secret,
    SecretManagerClient,
    SignifyClient,
)


class TestSignifySecretRotator(unittest.TestCase):
    """Tests for the Signify Secret Rotator.
    Tests the functionality of the Signify Secret Rotator, including the
    validation of messages, the rotation of secrets, and the handling of
    errors. It mocks the Google Cloud Secret Manager and Signify API to
    simulate the behavior of these services.
    """

    def setUp(self):
        """Set up the test case."""
        # Mock the Google Cloud Secret Manager
        access_secret_patcher = patch.object(SecretManagerClient, "get_secret")
        add_secret_version_patcher = patch.object(
            SecretManagerClient, "add_secret_version"
        )
        self.mock_access_secret = access_secret_patcher.start()
        self.mock_add_secret_version = add_secret_version_patcher.start()

        # Mock the Signify client
        fetch_access_token_patcher = patch.object(SignifyClient, "fetch_access_token")
        fetch_new_certificate_patcher = patch.object(
            SignifyClient, "fetch_new_certificate"
        )

        self.mock_fetch_access_token = fetch_access_token_patcher.start()
        self.mock_fetch_new_certificate = fetch_new_certificate_patcher.start()

        # Mock the logger
        self.logger = patch("pylogger.logger.create_logger", autospec=True).start()

        # Mock the Flask request
        self.app = Flask(__name__)
        self.app.config["TESTING"] = True

    def tearDown(self):
        """Tear down the test case."""
        patch.stopall()

    def test_rotate_signify_secret_success(self):
        """Test the rotation of a Signify secret."""
        # Mock the response from the Secret Manager.
        self.mock_access_secret.return_value = {
            "name": "projects/test/secrets/test",
            "tokenURL": "https://example.com/token",
            "certServiceURL": "https://example.com/cert",
            "clientID": "test-client-id",
            "certData": base64.b64encode(b"test-cert").decode("utf-8"),
            "privateKeyData": base64.b64encode(b"test-key").decode("utf-8"),
        }

        # Mock the response from the Signify client.
        self.mock_fetch_access_token.return_value = "test-access-token"
        self.mock_fetch_new_certificate.return_value = mocked_cert_create

        # Mock the request data.
        mock_pubsub_message = {
            "name": "projects/test/secrets/test",
            "labels": {
                "type": "signify",
                "usage": "tests",
                "owner": "test-owner",
            },
        }
        event = {
            "message": {
                "attributes": {
                    "eventType": "SECRET_ROTATE",
                },
                "data": base64.b64encode(
                    json.dumps(mock_pubsub_message).encode("utf-8")
                ).decode(),
            }
        }

        # Call the function to rotate the secret.
        with self.app.test_request_context(
            path="/",
            method="POST",
            json=event,
        ):
            response = rotate_signify_secret()

        # Check that the response is successful.
        self.assertEqual(response.status_code, 200)

    def test_rotate_signify_secret_unsupported_event_type(self):
        """Test the rotation of a Signify secret with an unsupported event type."""
        # Mock the request data.
        mock_pubsub_message = {
            "name": "projects/test/secrets/test",
            "labels": {
                "type": "signify",
                "usage": "tests",
                "owner": "test-owner",
            },
        }
        event = {
            "message": {
                "attributes": {
                    "eventType": "INVALID_EVENT",
                },
                "data": base64.b64encode(
                    json.dumps(mock_pubsub_message).encode("utf-8")
                ).decode(),
            }
        }

        # Call the function to rotate the secret.
        with self.app.test_request_context(
            path="/",
            method="POST",
            json=event,
        ):
            response = rotate_signify_secret()

        # Check that the response is an error.
        self.assertEqual(response.status_code, 500)

    def test_rotate_signify_secret_unsupported_incorrect_secret_type(self):
        """Test the rotation of a Signify secret with an incorrect secret type."""
        # Mock the request data.
        mock_pubsub_message = {
            "name": "projects/test/secrets/test",
            "labels": {
                "type": "service-account",
                "usage": "tests",
                "owner": "test-owner",
            },
        }
        event = {
            "message": {
                "attributes": {
                    "eventType": "SECRET_ROTATE",
                },
                "data": base64.b64encode(
                    json.dumps(mock_pubsub_message).encode("utf-8")
                ).decode(),
            }
        }

        # Call the function to rotate the secret.
        with self.app.test_request_context(
            path="/",
            method="POST",
            json=event,
        ):
            response = rotate_signify_secret()

        # Check that the response is an error.
        self.assertEqual(response.status_code, 200)

    def test_rotate_signify_secret_invalid_message(self):
        """Test the rotation of a Signify secret with an invalid message."""
        # Mock the request data.
        mock_pubsub_message = {}
        event = {
            "message": {
                "data": base64.b64encode(
                    json.dumps(mock_pubsub_message).encode("utf-8")
                ).decode(),
            }
        }

        # Call the function to rotate the secret.
        with self.app.test_request_context(
            path="/",
            method="POST",
            json=event,
        ):
            response = rotate_signify_secret()

        # Check that the response is an error.
        self.assertEqual(response.status_code, 500)


mocked_cert_create: List[x509.Certificate] = [
    x509.load_pem_x509_certificate(
        base64.b64decode(
            # pylint: disable=line-too-long
            "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUY3ekNDQTllZ0F3SUJBZ0lVYUlqenlHUytCUFFJZ0thYlhoU2xhTTk0dGpNd0RRWUpLb1pJaHZjTkFRRUwKQlFBd2dZWXhDekFKQmdOVkJBWVRBbGhZTVJJd0VBWURWUVFJREFsVGRHRjBaVTVoYldVeEVUQVBCZ05WQkFjTQpDRU5wZEhsT1lXMWxNUlF3RWdZRFZRUUtEQXREYjIxd1lXNTVUbUZ0WlRFYk1Ca0dBMVVFQ3d3U1EyOXRjR0Z1CmVWTmxZM1JwYjI1T1lXMWxNUjB3R3dZRFZRUUREQlJEYjIxdGIyNU9ZVzFsVDNKSWIzTjBibUZ0WlRBZUZ3MHkKTkRBNU1USXdPRFUwTWpkYUZ3MHpOREE1TVRBd09EVTBNamRhTUlHR01Rc3dDUVlEVlFRR0V3SllXREVTTUJBRwpBMVVFQ0F3SlUzUmhkR1ZPWVcxbE1SRXdEd1lEVlFRSERBaERhWFI1VG1GdFpURVVNQklHQTFVRUNnd0xRMjl0CmNHRnVlVTVoYldVeEd6QVpCZ05WQkFzTUVrTnZiWEJoYm5sVFpXTjBhVzl1VG1GdFpURWRNQnNHQTFVRUF3d1UKUTI5dGJXOXVUbUZ0WlU5eVNHOXpkRzVoYldVd2dnSWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUNEd0F3Z2dJSwpBb0lDQVFEa01hSWJzREhKSXBmZ3lOeUp2UWUycXZTS1ZhOWhSWGQ3SFB2SG1nZjJxUmkyejN6YmVRc0xuWnk1CmhkOE1LZnVHR25kb0FwdC9SR3RYRmgzTlowcVU2R1pIL2UxNWZWQmpqZW8rd3NtWHRXVmUrMGZuTE5QWXByT0IKdnFhZmYybUZmaytucUJCWE9XRExHV2FEclNrR0gvZ202WnZVOWNhNDh6YlRiOE0rRVc4T3VuWTRyTTZJUEcxNQpxSGJlRjRhMFF2WFJmbSswMytCbEJlUHFjWjg1ZEpjNnkvd1Qyajc0ckEyS1RLdWdrczUyUitGc0hOYS9jbEpTCko5SmlsOEdOa1RObmtRaGErZDZRcVpyUzBuTW4vSXJvdDlBdDN2Z0FWV2Qrb29DVWNxTUFxdURiVE1jRzhYUi8KcHI4bkl5bk5KVFhzYnJsSVZQL1BOaGdhZ09ra0gzYVFzVTIzMy9RUTFCNWlaMUpqMjhaSi9ycjBKZDV4TlJncwpoS3YzNUZxdzJ3SWFqT0tuWFBYaElHS2J0b0c0N09FSEJabGJvTTV6bHpqSCtPVERLSXY2NGg1dTBoWEt5ckZ1CjU5RTBTUFZsMEprYjRITVdtMEtzbGRiSjIzYUVsbmNIdlVBZjVWek01STlhMVVEYThjRW9RdTUrb0thaWVoaW4KTGE2QmNxUHM0R3pRQUdiTEpLdStTYk83SDR4aFJRanN3blY3WWkyVHJFbytiWDFIZXBhZ3oyL05MMS9iSlNSTgpNVmlmOGhRMFRaaHlxY1VGSWFhYmFyVldGb2RZVWFxeFRuUHIzdUM5L2pZak9Vd0tOSG9ySWhqejBIR005cHY3CjRUT0NmR01TY1RYQWo4Y1RqR0RUTjladjRQMGMrOFVycEx0ZjJOVldNUGtIczh3QzF3SURBUUFCbzFNd1VUQWQKQmdOVkhRNEVGZ1FVRHhtQmdjSVJpbVRDaGRkMHBoYnhBcjhZM0s4d0h3WURWUjBqQkJnd0ZvQVVEeG1CZ2NJUgppbVRDaGRkMHBoYnhBcjhZM0s4d0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHOXcwQkFRc0ZBQU9DCkFnRUFKazdkYXJISXRtUjVaK05qVk9WV1ZjNWtOMnVGRVdFbEkrazdHbjdEZDhRdE56VlBuQW5aK1IzSVdXQUEKSXdPQ2Z3YmJTZGRWakwxSnp3aVk5cFNhd3FEd3NlM1kzR3lzNUhHMXFqeFFCaFF3NXdVMTc4aHl6cGhwNkY1cwp5SmZxMCtvZ2hPdFNxOUp6bGFYNW5MdXJNZ3hWajBrZmNodXRoNjNFdnBwa29sNHcvb3Z3OTZvdEdUVmcvUm14Cm9yYnN4MVE5aERSalNNRUdaRHd2R0xONUx1K0JDQUFXcFp4N3c1VTl0bE9CaGxFSFF0cGl6cy9aZ1k5S0pabXAKc3dKMm0rTWp5K1RrMWt5ZW1nT3hmUDBZQkZDRnFHZU8xOWxqWjY1SmRTOUlSdnY5MEp5c1hHY0ZJbVkxRTVERQorTXovdkFaRVN4ZmRrSWFqN0xuZGZVNUJITVY0bDRDR0ZaRDdMYWpLcFJybzlybmU5QmRINHkrOVd0R1RHRlZCCmFzOVE3VGZDRTE1YkpYWE5QYzJXL0VQM1RjVTdlejI1WFcyR1JyVHAxVXA0MlhmRCtVUm9CUEdnU0lWSE4vdFoKbWxZbTlrQW8xajdKM2RnV3VlU1lFamxhRG1OSGlBbGtWOTJnSm55aDVvK3cxU1BudnZtMEkxejNyM0NVT25qYgp2UGFmVWE3TjdYcE1QREs5NW5rc2RZYkVlUWEzUDZ6V3VuL0VQUjdJblV1VVFBWFpPZUlEakVkTFZObzZqcDJuCjFpQUtlT3dNQU1FU00rWWZmSHVSN0FJY3ZHUXNBUU9hSlRjQVVjeldzSHFaR0RtRlVVU0tYSTBnL1ZyUFhsMVUKb0dBQ3pzNnZCbDZHaHo2RURhaURJVVlHTXM3em0yVjY0WTVRcGVJdTIvclJFWjQ9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
        )
    )
]
