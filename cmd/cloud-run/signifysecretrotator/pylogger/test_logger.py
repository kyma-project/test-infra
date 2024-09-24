"""Tests for logger module"""

import time
from typing import Any
import unittest
import logging
import json
from unittest.mock import Mock

from flask import Request

# pylint: disable=import-error
# False positive see: https://github.com/pylint-dev/pylint/issues/3984
from logger import GoogleCloudFormatter, create_logger


class TestGoogleCloudFormatter(unittest.TestCase):
    """Tests for custom formatter"""

    def setUp(self):
        self.component_name = "test_component"
        self.application_name = "test_application"
        self.log_fields = {"key1": "value1"}
        self.formatter = GoogleCloudFormatter(
            self.component_name, self.application_name, self.log_fields
        )

    def test_format(self) -> None:
        """Test format function"""
        log_record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="test_path",
            lineno=10,
            msg="This is a test message",
            args=(),
            exc_info=None,
        )
        formatted_log = self.formatter.format(log_record)
        expected_log = {
            "timestamp": log_record.created,
            "severity": "INFO",
            "message": "This is a test message",
        }
        self.assertEqual(json.loads(formatted_log), expected_log)


class TestCreateLogger(unittest.TestCase):
    """Tests for logger factory"""

    def setUp(self):
        # Ensure that each test has logger with unique name
        self.component_name = f"test_component_{time.time()}"
        self.application_name = "test_application"
        self.project_id = "test_project_id"
        self.request = Mock(spec=Request)
        self.request.headers = {"X-Cloud-Trace-Context": "1234567890/other-info"}

    def test_create_logger(self):
        """Tests create logger with trace"""

        logger: logging.Logger = create_logger(
            component_name=self.component_name,
            application_name=self.application_name,
            project_id=self.project_id,
            request=self.request,
            log_level=logging.INFO,
        )

        self.assertFalse(logger.isEnabledFor(logging.DEBUG))

        self.assertTrue(logger.hasHandlers())

        handler = logger.handlers[0]
        self.assertIsInstance(handler.formatter, GoogleCloudFormatter)

        expected_log_fields: dict[str, Any] = {
            "component": self.component_name,
            "labels": {"io.kyma.component": self.application_name},
            "logging.googleapi.com/trace": f"projects/{self.project_id}/traces/1234567890",
        }
        self.assertDictEqual(expected_log_fields, handler.formatter.log_fields)

    def test_create_logger_without_trace(self):
        """Tests create logger without request"""
        logger: logging.Logger = create_logger(
            component_name=self.component_name,
            application_name=self.application_name,
            log_level=logging.INFO,
        )

        self.assertFalse(logger.isEnabledFor(logging.DEBUG))
        self.assertTrue(logger.hasHandlers())

        handler = logger.handlers[0]
        self.assertIsInstance(handler.formatter, GoogleCloudFormatter)

        expected_log_fields: dict[str, Any] = {
            "component": self.component_name,
            "labels": {"io.kyma.component": self.application_name},
        }
        self.assertDictEqual(expected_log_fields, handler.formatter.log_fields)


if __name__ == "__main__":
    unittest.main()
