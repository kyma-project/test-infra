"""Module containing tests for signify secret rotator"""

import base64
from datetime import datetime, timedelta
import json
from typing import Any, Dict
import unittest
from unittest import mock
from unittest.mock import MagicMock, Mock, patch
from flask.testing import FlaskClient
from cryptography.hazmat.primitives import serialization, hashes
from cryptography import x509
from cryptography.x509.oid import NameOID
from cryptography.hazmat.primitives.asymmetric import rsa
import test_fixtures
from rotate_signify_certificate import (
    app,
    decrypt_private_key,
    extract_message_data,
    fetch_access_token,
    fetch_new_certificate,
    get_pubsub_message,
    prepare_log_fields,
    prepare_new_secret,
)

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


@patch("rotate_signify_certificate.project_id", project_id)
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


class TestFetchNewCertificate(unittest.TestCase):
    """Test cases for fetch_new_certificate function"""

    def setUp(self):
        """Set up test data"""
        self.cert_data = base64.b64decode(test_fixtures.mocked_secret_data["certData"])
        self.private_key = rsa.generate_private_key(
            public_exponent=65537,
            key_size=2048,
        )
        self.access_token = "test_token"
        self.certificate_service_url = "http://example.com/certificates"

    @patch("requests.post")
    def test_fetch_new_certificate(self, mock_post):
        """Test fetch_new_certificate success scenario"""
        # Create a mock response
        mock_response = MagicMock()
        mock_response.json.return_value = test_fixtures.mocked_cert_create_response
        mock_post.return_value = mock_response

        # Call the function with the test data
        certs = fetch_new_certificate(
            self.cert_data,
            self.private_key,
            self.access_token,
            self.certificate_service_url,
        )

        # Verify the mock post request
        mock_post.assert_called_once_with(
            self.certificate_service_url,
            headers={
                "Authorization": f"Bearer {self.access_token}",
                "Content-Type": "application/json",
                "Accept": "application/json",
            },
            data=json.dumps(
                {
                    "csr": {
                        "value": x509.CertificateSigningRequestBuilder()
                        .subject_name(
                            x509.load_pem_x509_certificate(self.cert_data).subject
                        )
                        .sign(self.private_key, hashes.SHA256())
                        .public_bytes(serialization.Encoding.PEM)
                        .decode("utf-8")
                    },
                    "validity": {"value": 7, "type": "DAYS"},
                    "policy": "sap-cloud-platform-clients",
                }
            ),
            timeout=10,
        )

        # Check that the expected number of certificates is returned
        self.assertEqual(len(certs), 1)


class TestPrepareNewSecret(unittest.TestCase):
    """Tests cases for prepare_new_secret function"""

    @staticmethod
    def generate_fake_cert_and_key():
        """
        Generates a fake RSA private key and a self-signed X.509 certificate.

        Returns:
            Tuple[x509.Certificate, rsa.RSAPrivateKey]: Tuple contains certificate and private key.
        """
        # pylint: disable=invalid-name
        ONE_DAY = timedelta(1, 0, 0)

        # Generate private key
        key = rsa.generate_private_key(
            public_exponent=65537,
            key_size=2048,
        )

        # Generate certificates
        subject = issuer = x509.Name(
            [
                x509.NameAttribute(NameOID.COUNTRY_NAME, "US"),
                x509.NameAttribute(NameOID.STATE_OR_PROVINCE_NAME, "California"),
                x509.NameAttribute(NameOID.LOCALITY_NAME, "San Francisco"),
                x509.NameAttribute(NameOID.ORGANIZATION_NAME, "My Company"),
                x509.NameAttribute(NameOID.COMMON_NAME, "mysite.com"),
            ]
        )
        cert = (
            x509.CertificateBuilder()
            .subject_name(subject)
            .issuer_name(issuer)
            .public_key(key.public_key())
            .serial_number(x509.random_serial_number())
            .not_valid_before(datetime.now())
            .not_valid_after(datetime.now() + ONE_DAY)
            .sign(key, hashes.SHA256())
        )

        return cert, key

    def setUp(self):
        self.cert, self.key = self.generate_fake_cert_and_key()
        self.certificates = [self.cert]
        self.private_key = self.key
        self.secret_data = {
            "certServiceURL": "https://cert.example.com",
            "tokenURL": "https://token.example.com",
            "clientID": "example-client-id",
        }
        self.created_at = "2023-01-01T00:00:00Z"

    def test_prepare_new_secret_structure(self):
        """Test if the output structure of `prepare_new_secret` matches the expected format"""
        result = prepare_new_secret(
            self.certificates, self.private_key, self.secret_data, self.created_at
        )

        self.assertIn("certData", result)
        self.assertIn("privateKeyData", result)
        self.assertIn("createdAt", result)
        self.assertIn("certServiceURL", result)
        self.assertIn("tokenURL", result)
        self.assertIn("clientID", result)
        self.assertIn("password", result)

        self.assertEqual(result["createdAt"], self.created_at)
        self.assertEqual(result["certServiceURL"], self.secret_data["certServiceURL"])
        self.assertEqual(result["tokenURL"], self.secret_data["tokenURL"])
        self.assertEqual(result["clientID"], self.secret_data["clientID"])
        self.assertEqual(result["password"], "")

    def test_prepare_new_secret_cert_data(self):
        """Test if the certificate data in the output is correctly formatted and base64 encoded"""
        result = prepare_new_secret(
            self.certificates, self.private_key, self.secret_data, self.created_at
        )
        encoded_cert_data = result["certData"]
        cert_data_decoded = base64.b64decode(encoded_cert_data).decode()

        cert_subject = f"subject={self.cert.subject.rfc4514_string()}\n"
        cert_issuer = f"issuer={self.cert.issuer.rfc4514_string()}\n"
        cert_body = f"{self.cert.public_bytes(serialization.Encoding.PEM).decode()}\n"

        self.assertIn(cert_subject, cert_data_decoded)
        self.assertIn(cert_issuer, cert_data_decoded)
        self.assertIn(cert_body, cert_data_decoded)

    def test_prepare_new_secret_private_key_data(self):
        """Test if the private key data in the output is correctly formatted and base64 encoded"""
        result = prepare_new_secret(
            self.certificates, self.private_key, self.secret_data, self.created_at
        )
        encoded_private_key_data = result["privateKeyData"]
        private_key_data_decoded = base64.b64decode(encoded_private_key_data)

        self.assertEqual(
            private_key_data_decoded,
            self.private_key.private_bytes(
                serialization.Encoding.PEM,
                serialization.PrivateFormat.PKCS8,
                serialization.NoEncryption(),
            ),
        )


class TestExtractMessageData(unittest.TestCase):
    """Unit tests for the extract_message_data function."""

    def test_valid_message(self):
        """Test a valid Pub/Sub message with correct event type and encoded data."""
        data = {"key": "value"}
        encoded_data = base64.b64encode(json.dumps(data).encode()).decode()
        pubsub_message = {
            "data": encoded_data,
            "attributes": {"eventType": "SECRET_ROTATE"},
        }

        result = extract_message_data(pubsub_message)
        self.assertEqual(result, data)

    def test_unsupported_event_type(self):
        """Test with an unsupported event type to ensure ValueError is raised."""
        pubsub_message = {
            "data": base64.b64encode(json.dumps({"key": "value"}).encode()).decode(),
            "attributes": {"eventType": "UNSUPPORTED_EVENT"},
        }

        with self.assertRaises(ValueError) as context:
            extract_message_data(pubsub_message)

        self.assertEqual(str(context.exception), "Unsupported event type")

    def test_no_attributes(self):
        """Test with no attributes to ensure ValueError is raised."""
        pubsub_message = {
            "data": base64.b64encode(json.dumps({"key": "value"}).encode()).decode()
        }

        with self.assertRaises(ValueError) as context:
            extract_message_data(pubsub_message)

        self.assertEqual(str(context.exception), "Unsupported event type")

    def test_invalid_base64_data(self):
        """Test with invalid base64-encoded data to ensure an exception is raised."""
        pubsub_message = {
            "data": "invalid_base64_data",
            "attributes": {"eventType": "SECRET_ROTATE"},
        }

        with self.assertRaises(Exception):
            extract_message_data(pubsub_message)


class TestDecryptPrivateKey(unittest.TestCase):
    """Unit tests for decrypt_private_key function."""

    def setUp(self):
        """Set up test data for the tests."""
        # Generate private key
        self.private_key = rsa.generate_private_key(
            public_exponent=65537,
            key_size=2048,
        )

        self.password = b"password"

        # Encrypt private key with the password
        self.private_key_data = self.private_key.private_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PrivateFormat.TraditionalOpenSSL,
            encryption_algorithm=serialization.BestAvailableEncryption(self.password),
        )

    def test_decrypt_private_key(self):
        """Test the decryption of an encrypted private key."""
        # Decrypt the private key
        decrypted_private_key = decrypt_private_key(
            self.private_key_data, self.password
        )

        # Convert the decrypted private key back to an object for validation
        decrypted_private_key_obj = serialization.load_pem_private_key(
            decrypted_private_key, password=None
        )

        # Serialize the private key object with no encryption to compare
        expected_private_key_data = decrypted_private_key_obj.private_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PrivateFormat.PKCS8,
            encryption_algorithm=serialization.NoEncryption(),
        )

        self.assertEqual(decrypted_private_key, expected_private_key_data)
        self.assertTrue(
            decrypted_private_key.startswith(b"-----BEGIN PRIVATE KEY-----")
        )
        self.assertNotIn(b"ENCRYPTED", decrypted_private_key)


class TestFetchAccessToken(unittest.TestCase):
    """Test case for the fetch_access_token function."""

    @patch("requests.post")
    def test_fetch_access_token(self, mock_post):
        """Test the fetch_access_token function."""
        certificate = b"dummy_certificate"
        private_key = b"dummy_private_key"
        token_url = "https://dummy_token_url"
        client_id = "dummy_client_id"

        expected_token = "dummy_access_token"

        # Setup the mock response
        mock_response = unittest.mock.Mock()
        mock_response.json.return_value = {"access_token": expected_token}
        mock_post.return_value = mock_response

        # Call the function
        result = fetch_access_token(certificate, private_key, token_url, client_id)

        # Assert the result
        self.assertEqual(result, expected_token)

        # Assert that post was called with the expected parameters
        self.assertTrue(mock_post.called)
        self.assertEqual(
            mock_post.call_args[1]["cert"], (mock.ANY, mock.ANY)
        )  # file paths are temporary
        self.assertEqual(
            mock_post.call_args[1]["data"]["grant_type"], "client_credentials"
        )
        self.assertEqual(mock_post.call_args[1]["data"]["client_id"], client_id)
        self.assertEqual(mock_post.call_args[1]["timeout"], 30)


if __name__ == "__main__":
    unittest.main()
