"""Encapsulates logging for cloud runs"""

import json
import logging
from typing import Any, Dict

from flask import Request


class GoogleCloudFormatter(logging.Formatter):
    """Wraps the formatting the logs using google cloud format"""

    def __init__(
        self, component_name: str, application_name: str, log_fields: Dict[str, Any]
    ) -> None:
        self.component_name: str = component_name
        self.application_name: str = application_name
        self.log_fields: Dict[str, Any] = log_fields

        super().__init__()

    def format(self, record: logging.LogRecord) -> str:
        """Formats record into cloud event log"""

        return json.dumps(
            {
                "timestamp": record.created,
                "severity": record.levelname,
                "message": record.getMessage(),
            }
        )


def create_logger(
    component_name: str,
    application_name: str,
    project_id: str = None,
    request: Request = None,
    log_level=logging.INFO,
) -> logging.Logger:
    """Creates instance of stdout logger for aplication's component"""
    logger: logging.Logger = logging.getLogger(f"{application_name}/{component_name}")
    logger.setLevel(log_level)

    log_fields = {
        "component": component_name,
        "labels": {"io.kyma.component": application_name},
    }

    if request:
        trace_header: str | None = request.headers.get("X-Cloud-Trace-Context")

        if trace_header and project_id:
            trace: list[str] = trace_header.split("/")
            log_fields["logging.googleapi.com/trace"] = (
                f"projects/{project_id}/traces/{trace[0]}"
            )

    formatter = GoogleCloudFormatter(
        component_name=component_name,
        application_name=application_name,
        log_fields=log_fields,
    )
    handler = logging.StreamHandler()
    handler.setFormatter(formatter)

    logger.addHandler(handler)

    return logger
