"""Custom wrapper for Google's Secret Manager Service client"""

import json
from typing import Any, Dict
from google.cloud import secretmanager
from google.api_core.exceptions import GoogleAPIError


class SecretManagerClient:
    """
    Wraps the google's secret manager client implementation.
    Provides more efficient way to retrieve and set secrets in kyma-project secret manager
    """

    def __init__(self, client=secretmanager.SecretManagerServiceClient()) -> None:
        self.client: secretmanager.SecretManagerServiceClient = client

    def get_secret(
        self, secret_id: str, secret_version: str = "latest", is_json: bool = True
    ) -> Dict[str, Any]:
        """Fetches value of secret with given version

        Args:
            secret_id (str): Secret id in format "projects/<project_id>/secrets/<secret name>"
            secret_version (str, optional): Version of the secret. Defaults to "latest".
            is_json (bool): Secret is json struct. Defaults to True

        Returns:
            Dict[str, Any]: JSON decoded or str depending on is_json value
        """

        secret_name = f"{secret_id}/versions/{secret_version}"

        try:
            response: secretmanager.AccessSecretVersionResponse = (
                self.client.access_secret_version(name=secret_name)
            )
            secret_value = response.payload.data.decode()

            if is_json:
                return json.loads(secret_value)

            return secret_value
        except GoogleAPIError as e:
            raise SecretManagerError(secret_id, e) from e

    def add_secret_version(self, secret_id: str, data: str) -> None:
        """Adds new secret version with given data

        Args:
            secret_id (str): Secret id in format "projects/<project_id>/secrets/<secret name>"
            data (str): Value that should be set as new secret version
        """
        payload = {"data": data.encode()}

        try:
            self.client.add_secret_version(parent=secret_id, payload=payload)
        except GoogleAPIError as e:
            raise SecretManagerError(secret_id, e) from e


class SecretManagerError(Exception):
    """Common class for Secret Manager client exceptions"""

    def __init__(self, secret_id: str, e: Exception) -> None:
        self.add_note(f"Failed to access secret {secret_id}, error: {e}")
