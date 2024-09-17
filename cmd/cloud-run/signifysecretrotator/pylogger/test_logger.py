"""Tests for logger module"""

import unittest
from unittest.mock import patch
import json

# pylint: disable=import-error
# False positive see: https://github.com/pylint-dev/pylint/issues/3984
from logger import Logger, LogEntry


class TestLogger(unittest.TestCase):
    """Tests for logger class"""

    @patch("flask.Request")
    def test_logger_initialization(self, mock_request):
        """Test logger initialization"""
        mock_request = mock_request()
        mock_request.headers.get.return_value = "trace-id/other-info"

        logger = Logger("component1", "application1", "project1", mock_request)

        self.assertEqual(logger.log_fields["component"], "component1")
        self.assertEqual(
            logger.log_fields["labels"]["io.kyma.component"], "application1"
        )
        self.assertTrue("logging.googleapi.com/trace" in logger.log_fields)

    def test_logger_initialization_without_request(self):
        """Test logger initialization without flas request"""
        logger = Logger("component1", "application1")

        self.assertEqual(logger.log_fields["component"], "component1")
        self.assertEqual(
            logger.log_fields["labels"]["io.kyma.component"], "application1"
        )
        self.assertFalse("logging.googleapi.com/trace" in logger.log_fields)

    @patch("builtins.print")
    def test_log_error(self, mock_print):
        """Test log_error method"""
        logger = Logger("component1", "application1")
        logger.log_error("An error occurred")

        expected_output = LogEntry(
            severity="ERROR",
            message="An error occurred",
            component="component1",
            labels={"io.kyma.component": "application1"},
        )

        mock_print.assert_called_once_with(expected_output)

    @patch("builtins.print")
    def test_log_info(self, mock_print):
        """Test log_info method"""
        logger = Logger("component1", "application1")
        logger.log_info("Information message")

        expected_output = LogEntry(
            severity="INFO",
            message="Information message",
            component="component1",
            labels={"io.kyma.component": "application1"},
        )

        mock_print.assert_called_once_with(expected_output)

    def test_log_entry_str(self):
        """Test log entry to json"""
        log_entry = LogEntry(
            severity="INFO", message="Test", component="test_component"
        )
        expected_output = json.dumps(
            {"severity": "INFO", "message": "Test", "component": "test_component"}
        )
        self.assertEqual(str(log_entry), expected_output)


if __name__ == "__main__":
    unittest.main()
