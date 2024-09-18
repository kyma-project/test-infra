"""Tests for signify client module"""

import base64
import unittest
from unittest.mock import patch, MagicMock

import requests

# pylint: disable=import-error
# False positive see: https://github.com/pylint-dev/pylint/issues/3984
from client import SignifyClient
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.serialization import Encoding, PrivateFormat

# pylint: disable=import-error
# False positive see: https://github.com/pylint-dev/pylint/issues/3984
import test_fixtures


class TestSignifyClient(unittest.TestCase):
    """
    Unit tests for the SignifyClient class.
    """

    def setUp(self):
        """
        Set up method to initialize the SignifyClient object and necessary data for tests.
        """
        self.token_url = "https://example.com/token"
        self.certificate_service_url = "https://example.com/certificate"
        self.client = SignifyClient(
            token_url=self.token_url,
            certificate_service_url=self.certificate_service_url,
            client_id="fake_client_id",
        )
        self.certificate = base64.b64decode(
            test_fixtures.mocked_secret_data["certData"]
        )
        self.private_key = rsa.generate_private_key(
            public_exponent=65537, key_size=2048
        )
        self.access_token = "fake_access_token"

    @patch("requests.post")
    def test_fetch_access_token_success(self, mock_post):
        """
        Test successful fetch of access token.

        Mock the requests.post to return a successful response with the expected access token.
        """
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.json.return_value = {"access_token": self.access_token}
        mock_post.return_value = mock_response
        private_key_bytes = self.private_key.private_bytes(
            Encoding.PEM, PrivateFormat.PKCS8, serialization.NoEncryption()
        )

        token = self.client.fetch_access_token(
            self.certificate,
            private_key_bytes,
        )

        self.assertEqual(token, self.access_token)
        mock_post.assert_called_once()

    @patch("requests.post")
    def test_fetch_access_token_failed_status_code(self, mock_post):
        """
        Test fetch of access token when the response status code is not 200.
        """
        mock_response = MagicMock()
        mock_response.status_code = 400
        mock_response.json.return_value = {}
        mock_post.return_value = mock_response
        private_key_bytes = self.private_key.private_bytes(
            Encoding.PEM, PrivateFormat.PKCS8, serialization.NoEncryption()
        )

        with self.assertRaises(requests.HTTPError):
            self.client.fetch_access_token(
                certificate=self.certificate, private_key=private_key_bytes
            )
        mock_post.assert_called_once()

    @patch("requests.post")
    def test_fetch_access_token_unexpected_response(self, mock_post):
        """
        Test fetch of access token when the response does not contain the expected structure.
        """
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.json.return_value = {}
        mock_post.return_value = mock_response
        private_key_bytes = self.private_key.private_bytes(
            Encoding.PEM, PrivateFormat.PKCS8, serialization.NoEncryption()
        )

        with self.assertRaises(ValueError):
            self.client.fetch_access_token(
                certificate=self.certificate, private_key=private_key_bytes
            )
        mock_post.assert_called_once()

    @patch("requests.post")
    def test_fetch_new_certificate_success(self, mock_post):
        """
        Test successful fetch of a new certificate.
        """
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.json.return_value = test_fixtures.mocked_cert_create_response
        mock_post.return_value = mock_response

        certs = self.client.fetch_new_certificate(
            self.certificate, self.private_key, self.access_token
        )

        self.assertEqual(len(certs), 1)
        mock_post.assert_called_once()

    @patch("requests.post")
    def test_fetch_new_certificate_failed_status_code(self, mock_post):
        """
        Test fetch of a new certificate when the response status code is not 200.s
        """
        mock_response = MagicMock()
        mock_response.status_code = 400
        mock_response.json.return_value = {}
        mock_post.return_value = mock_response

        with self.assertRaises(KeyError):
            self.client.fetch_new_certificate(
                self.certificate, self.private_key, self.access_token
            )
        mock_post.assert_called_once()


if __name__ == "__main__":
    unittest.main()
