"""Package that contains the message validator for the Signify Secret Rotator."""


class SecretTypeError(Exception):
    """Custom exception for invalid secret type."""

    def __init__(self, message: str, received_type: str) -> None:
        super().__init__(message)
        self.message = message
        self.received_type = received_type


class EventTypeError(Exception):
    """Custom exception for invalid event type."""

    def __init__(self, message: str, received_event_type: str) -> None:
        super().__init__(message)
        self.message = message
        self.received_event_type = received_event_type


# MessageValidator class verifies the message type and event type
# it don't require few public methods, one is sufficient
# Ignoring pylint error for too few public methods
# pylint: disable=too-few-public-methods
class MessageValidator:
    """Class that validates the message received from Pub/Sub.
    It checks if the secret type is valid based on the provided secret type."""

    def __init__(self, valid_secret_type: str) -> None:
        self.valid_secret_type = valid_secret_type

    def validate(self, message: dict) -> None:
        """
        Validates the message received from Pub/Sub.
        Raises an exception when message is invalid
        """

        self._verify_secret_type(message)

    def _verify_secret_type(self, message: dict) -> None:
        """
        Verifies if the secret type is valid based on provied secret type.
        Raises an exception when secret type is invalid.
        Secret type is defined as a type label in the secret manager
        """

        secret_type = message.get("labels", {}).get("type")
        if secret_type != self.valid_secret_type:
            raise SecretTypeError(
                f"Secret type {secret_type} is not valid. Expected {self.valid_secret_type}",
                secret_type,
            )
