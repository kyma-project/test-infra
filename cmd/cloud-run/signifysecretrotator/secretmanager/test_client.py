"""Contains tests for secret manager client module"""

import json
import unittest
from unittest.mock import MagicMock

from secretmanager.client import SecretManagerClient


class TestSecretManagerClient(unittest.TestCase):
    """Tests for secret manager client"""

    def setUp(self) -> None:
        # Create a mock client instance
        self.mock_client = MagicMock()
        self.client = SecretManagerClient(client=self.mock_client)

    def test_get_secret_json(self) -> None:
        """Tests fetching json secret data"""
        # Arrange
        mock_response = MagicMock()
        mock_response.payload.data.decode.return_value = json.dumps({"key": "value"})
        self.mock_client.access_secret_version.return_value = mock_response

        # Act
        secret = self.client.get_secret("projects/test-project/secrets/test-secret")

        # Assert
        self.assertEqual(secret, {"key": "value"})
        self.mock_client.access_secret_version.assert_called_once_with(
            name="projects/test-project/secrets/test-secret/versions/latest"
        )

    def test_get_secret_plain_string(self) -> None:
        """Tests fetching string secret data"""
        # Arrange
        mock_response = MagicMock()
        mock_response.payload.data.decode.return_value = "some-secret-value"
        self.mock_client.access_secret_version.return_value = mock_response

        # Act
        secret = self.client.get_secret(
            "projects/test-project/secrets/test-secret", is_json=False
        )

        # Assert
        self.assertEqual(secret, "some-secret-value")
        self.mock_client.access_secret_version.assert_called_once_with(
            name="projects/test-project/secrets/test-secret/versions/latest"
        )

    def test_add_secret_version(self) -> None:
        """Tests setting a new secret version"""
        # Arrange
        secret_id = "projects/test-project/secrets/test-secret"
        secret_data = "new-secret-value"

        # Act
        self.client.add_secret_version(secret_id, secret_data)

        # Assert
        payload: dict[str, bytes] = {"data": secret_data.encode()}
        self.mock_client.add_secret_version.assert_called_once_with(
            parent=secret_id, payload=payload
        )


if __name__ == "__main__":
    unittest.main()
