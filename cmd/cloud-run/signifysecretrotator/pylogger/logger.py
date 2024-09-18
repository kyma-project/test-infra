"""Encapsulates logging for cloud runs"""

import json
from typing import Any, Dict

from flask import Request


class LogEntry(dict):
    """Simplifies logging by returning a JSON string."""

    def __str__(self) -> str:
        return json.dumps(self)


class Logger:
    """Encapsulates logging for cloud runs"""

    def __init__(
        self,
        component_name: str,
        application_name: str,
        project_id: str = "",
        request: Request = None,
    ) -> None:
        self.log_fields: Dict[str, Any] = {
            "component": component_name,
            "labels": {"io.kyma.component": application_name},
        }
        if request:
            trace_header: str = request.headers.get("X-Cloud-Trace-Context")

            if trace_header and project_id:
                trace = trace_header.split("/")
                self.log_fields["logging.googleapi.com/trace"] = (
                    f"projects/{project_id}/traces/{trace[0]}"
                )

    def log_error(self, message: str) -> None:
        """Print log error message

        Args:
            message (str): Custom message that should be placed in entry
        """
        self._log(message, "ERROR")

    def log_info(self, message: str) -> None:
        """Print log info message

        Args:
            message (str): Custom message that should be placed in entry
        """
        self._log(message, "INFO")

    def _log(self, message: str, severity: str) -> None:
        entry = LogEntry(severity=severity, message=message, **self.log_fields)

        print(entry)
