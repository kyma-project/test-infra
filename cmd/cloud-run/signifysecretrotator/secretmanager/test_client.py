"""Contains tests for secret manager client module"""

import json
import unittest
from unittest.mock import AsyncMock, MagicMock, patch
from google.cloud import secretmanager

# pylint: disable=import-error
# False positive see: https://github.com/pylint-dev/pylint/issues/3984
from client import SecretManagerClient


class TestSecretManagerClient(unittest.TestCase):
    """Tests for secret manager client"""

    def setUp(self) -> None:
        access_secret_patcher = patch.object(
            secretmanager.SecretManagerServiceClient, "access_secret_version"
        )
        add_secret_version_patcher = patch.object(
            secretmanager.SecretManagerServiceClient, "add_secret_version"
        )

        self.mock_access_secret_version: MagicMock | AsyncMock = (
            access_secret_patcher.start()
        )
        self.mock_add_secret_version: MagicMock | AsyncMock = (
            add_secret_version_patcher.start()
        )

        self.addCleanup(access_secret_patcher.stop)
        self.addCleanup(add_secret_version_patcher.stop)

        self.client = SecretManagerClient()

    def test_get_secret_json(self) -> None:
        """Tests fetching json secret data"""
        # Arrange
        mock_response = MagicMock()
        mock_response.payload.data.decode.return_value = json.dumps({"key": "value"})
        self.mock_access_secret_version.return_value = mock_response

        # Act
        secret = self.client.get_secret("projects/test-project/secrets/test-secret")

        # Assert
        self.assertEqual(secret, {"key": "value"})
        self.mock_access_secret_version.assert_called_once_with(
            secret_name="projects/test-project/secrets/test-secret/versions/latest"
        )

    def test_get_secret_plain_string(self) -> None:
        """Tests fetching string secret data"""
        # Arrange
        mock_response = MagicMock()
        mock_response.payload.data.decode.return_value = "some-secret-value"
        self.mock_access_secret_version.return_value = mock_response

        # Act
        secret = self.client.get_secret(
            "projects/test-project/secrets/test-secret", is_json=False
        )

        # Assert
        self.assertEqual(secret, "some-secret-value")
        self.mock_access_secret_version.assert_called_once_with(
            secret_name="projects/test-project/secrets/test-secret/versions/latest"
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
        self.mock_add_secret_version.assert_called_once_with(
            parent=secret_id, payload=payload
        )


if __name__ == "__main__":
    unittest.main()
