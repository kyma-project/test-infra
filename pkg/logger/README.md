# Logger

## Overview

`pkg/logger` is a structured logging library for GCP-based workloads. It provides a unified logging interface that writes GCP-compatible structured JSON to console (stdout/stderr), directly to the Cloud Logging API, or both — controlled by the **LOG_DESTINATION** environment variable.

The package satisfies ADR-006 and replaces direct usage of `go.uber.org/zap` across the codebase.

## Prerequisites

- Go 1.26 or later
- A GCP project with the Cloud Logging API enabled when using `api` or `console-and-api` destination
- [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials) configured when using `api` or `console-and-api` destination

## Usage

### Basic Setup

Initialize the logger using the `New()` factory function. Configure it using environment variables — no code changes are required between environments.

```go
import "github.com/kyma-project/test-infra/pkg/logger"

log, err := logger.New()
if err != nil {
    panic(err)
}
defer log.Sync()

log.Infow("server started", "port", 8080)
```

### Adding Labels

Labels are indexed and filterable in Cloud Logging. Use them for static metadata such as application name, version, and environment. Use regular key-value pairs for dynamic, per-request data.

```go
log, err := logger.New()
if err != nil {
    panic(err)
}
defer log.Sync()

log = log.With(
    logger.LogLabel("app", "image-builder"),
    logger.LogLabel("version", "1.2.0"),
    logger.LogLabel("environment", "production"),
)

log.Infow("handling request", "request_id", "abc-123")
```

> [!NOTE]
> ADR-006 requires the following labels on all log entries: `app`, `version`, `environment`. Add them using `logger.LogLabel()` during logger initialization.

### Child Loggers

Use `With()` to create a child logger that includes additional fields in every subsequent log entry.

```go
log, err := logger.New()
if err != nil {
    panic(err)
}
defer log.Sync()

requestLog := log.With("request_id", "abc-123", "user_id", "42")
requestLog.Infow("processing request")
requestLog.Infow("request completed", "status", 200)
```

## Configuration

| Variable | Required | Default | Description |
|---|---|---|---|
| **LOG_DESTINATION** | No | `console` | Where to send logs. See [Log Destinations](#log-destinations). |
| **LOG_LEVEL** | No | `info` | Minimum severity: `debug`, `info`, `warn`, `error`, `dpanic`, `panic`, `fatal`. |
| **GCP_PROJECT_ID** | Conditional | — | GCP project ID. Required when **LOG_DESTINATION** is `api` or `console-and-api`. |
| **GCP_LOG_NAME** | No | `application` | Log name in Cloud Logging. |


### Log Destinations

| Value | Behavior |
|---|---|
| `console` | Writes structured JSON to stdout/stderr. Use on Cloud Run and GKE — the agent collects stdout automatically. |
| `api` | Sends logs directly to the Cloud Logging API. |
| `console-and-api` | Writes to both stdout/stderr and Cloud Logging API simultaneously. |

## Authentication Outside GCP

When **LOG_DESTINATION** is `api` or `console-and-api`, the logger requires GCP credentials. Credentials and project access are validated at startup — if the project does not exist or the credentials lack the **roles/logging.logWriter** role, `New()` returns an error immediately.


**Local development:**
```bash
gcloud auth application-default login
```

**Container outside GCP:** Mount a service account key with the **roles/logging.logWriter** role:
```bash
docker run \
  -v /path/to/sa-key.json:/tmp/sa-key.json \
  -e GOOGLE_APPLICATION_CREDENTIALS=/tmp/sa-key.json \
  -e LOG_DESTINATION=console-and-api \
  -e GCP_PROJECT_ID=your-project-id \
  your-image
```

> [!NOTE]
> The GCP client library refreshes credentials automatically. Long-running containers do not require a restart when using Workload Identity Federation with a stable credential source such as AWS or Azure instance metadata.

## Testing

Use `BufferLogger` in unit tests to capture log output without real I/O:

```go
buf := logger.NewBufferLogger()
buf.Infow("something happened", "key", "value")

entries := buf.Entries()
// entries[0].Message == "something happened"
```

## Log Format

All destinations produce the same JSON structure compatible with Cloud Logging structured log ingestion:

```json
{
  "severity": "INFO",
  "timestamp": "2026-04-10T17:49:21Z",
  "message": "server started",
  "port": 8080,
  "logging.googleapis.com/labels": {
    "app": "image-builder",
    "version": "1.2.0",
    "environment": "production"
  },
  "logging.googleapis.com/sourceLocation": {
    "file": "main.go",
    "line": 42,
    "function": "main.main"
  }
}
```

## Related Links

- [uber-go/zap](https://github.com/uber-go/zap)
- [pkg/logging](../logging/) — general-purpose logger (deprecated, use `pkg/logger` for new workloads)
- [pkg/gcp/logging](../gcp/logging/) — GCP API logging wrapper (deprecated, use `pkg/logger` for new workloads)
