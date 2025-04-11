"""Package for tests cases for message vlaidator class"""

import unittest
from messagevalidator import (
    MessageValidator,
    MessageTypeError,
)


class TestMessageValidator(unittest.TestCase):
    """Test cases for the MessageValidator class."""

    def test_raises_invalid_message_type_on_not_expected_message_type(self):
        """Test that an InvalidMessageTypeError is raised when the message type is not expected."""

        message = {"labels": {"type": "un"}}
        validator = MessageValidator("expected_type")

        with self.assertRaises(MessageTypeError):
            validator.validate(message)

    def test_pass_valid_message_type_without_error(self):
        """Test that no error is raised when the message type is valid."""

        message = {"labels": {"type": "expected_type"}}
        validator = MessageValidator("expected_type")

        try:
            validator.validate(message)
        except MessageTypeError:
            self.fail("MessageTypeError was raised unexpectedly!")
