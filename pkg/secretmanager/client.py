"""Custom wrapper for Google's Secret Manager Service client"""

from google.cloud.secretmanager import SecretManagerServiceClient


class SecretManagerClient:
    """
    Wraps the google's secret manager client implementation.
    Provides more efficient way to retrieve and set secrets in kyma-project secret manager
    """

    def __init__(self, client=SecretManagerServiceClient()) -> None:
        self.client: SecretManagerClient = client

    def get_secret(self, secret_id: str, secret_version: str = "latest"):
        """Gets given secret in given version"""
        return "Test"
