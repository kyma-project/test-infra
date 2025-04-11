"""Package that contains the message validator for the Signify Secret Rotator."""


class MessageTypeError(Exception):
    """Custom exception for invalid message type."""

    def __init__(self, message: str, received_type: str) -> None:
        super().__init__(message)
        self.message = message
        self.received_type = received_type


# MessageValidator class verifies the message type and event type
# pylint: disable=too-few-public-methods
class MessageValidator:
    """Class that validates the message received from Pub/Sub."""

    def __init__(self, valid_secret_type: str):
        self.valid_secret_type = valid_secret_type

    def validate(self, message: dict) -> None:
        """
        Validates the message received from Pub/Sub.
        Raises an exception when message is invalid
        """

        self._verify_message_type(message)

    def _verify_message_type(self, message: dict) -> None:
        """
        Verifies if the message type is valid.
        """

        message_type = message.get("labels", {}).get("type")
        if message_type != self.valid_secret_type:
            raise MessageTypeError(
                "Incorrect message type received. "
                + f"Expected {self.valid_secret_type}, got {message_type}",
                message_type,
            )
