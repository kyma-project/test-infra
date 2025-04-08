"""Tests for signifysecretrotator module"""

import unittest
import logging
from unittest.mock import Mock


from signifysecretrotator import validate_message


class TestValidateMessage(unittest.TestCase):
    """Tests for validate_message function"""

    def setUp(self):
        self.logger = Mock(spec=logging.Logger)
        self.secret_rotate_message_type = "signify"

    def test_validate_message_correct_type(self):
        """Test validate_message with correct message type"""
        message = {"labels": {"type": self.secret_rotate_message_type}}
        result = validate_message(self.logger, message)
        self.assertTrue(result)

    def test_validate_message_incorrect_type(self):
        """Test validate_message with incorrect message type"""
        message = {"labels": {"type": "incorrect_type"}}
        result = validate_message(self.logger, message)
        self.assertFalse(result)
        self.logger.info.assert_called_once_with(
            "Incorrect message type received. Expected %s, got %s. Ignoring message.",
            "incorrect_type",
            self.secret_rotate_message_type,
        )

    def test_validate_message_missing_labels(self):
        """Test validate_message with missing labels"""
        message = {}
        result = validate_message(self.logger, message)
        self.assertFalse(result)
        self.logger.info.assert_called_once_with(
            "Incorrect message type received. Expected %s, got %s. Ignoring message.",
            None,
            self.secret_rotate_message_type,
        )

    def test_validate_message_missing_type(self):
        """Test validate_message with missing type in labels"""
        message = {"labels": {}}
        result = validate_message(self.logger, message)
        self.assertFalse(result)
        self.logger.info.assert_called_once_with(
            "Incorrect message type received. Expected %s, got %s. Ignoring message.",
            None,
            self.secret_rotate_message_type,
        )


if __name__ == "__main__":
    unittest.main()
