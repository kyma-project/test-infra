# Logging

## Description

Logging is a critical part of any application. It provides a way to track the execution of the application and to debug issues.
It can serve as a source of information for monitoring and auditing purposes.
It helps during the development and testing phases of the application.
Logs are read by humans and machines, so it is important to have a consistent format and structure.

## Standards

In order to make sure our logs are useful and our tools are using them effectively, we define the following standards. They must be followed
by all developers in Neighbors team and implemented in all applications.

### Interfaces

All functions must accept a logging interface defined in the [logging](../../pkg/logging/types.go) package.
These interfaces define a set of methods to log messages at different levels.
The interfaces defines a methods to log using strig, formated strings and structured data.

### Logger Parameter

Logger parameter must be defined as the first parameter or immediately after the context parameter in all functions.

### Log Levels

All functions must log messages at the appropriate level. The following levels are minimum required:

- `Debug`: Detailed information about the application execution. This level can be used to log variable values and its transformation. This
  level must be disabled by default in production environments. It's used for development and troubleshooting purposes.
- `Info`: Informational messages about the application execution. This level can be used to log application state changes and important
  events. This level can be used to log input and output data. This level must be used to log normal and correct application behaviour and
  must be enabled by default in production environments.
- `Error`: Error messages about the application execution. This level must be used to log application errors and exceptions. It describes
  abnormal, unexpected and undesired application behaviour. This level must be enabled by default in production environments.

### Structured Logging

All logs must be structured.
This means that logs must be written as JSON, to allow easy parsing and analysis by machines.
It's allowed to log messages in plain text format, but structured data must be included in the log message.
It's allowed to use simple or formated string for debug logs.

### Structured Data

The structured data must include the following fields:

- application: the name of the application for non package level logs
- component: the name of the component that generated the log
- message: the log message
- data: the data associated with the log message, like function parameters, return values, and other relevant data
- trace: the trace id of the request

### Timestamps

All logs must include a timestamp. The timestamp must be in the ISO 8601 format with time offset to UTC suffix.

### Context

Each log message must describe the context in which it was generated with sufficient details.
The context can be described in the log message and in the structured data.
It must provide enough information to understand application state and behaviour.
It's allowed to assume previous log messages are available and required to describe the context.

### Tracing

### Common Implementation

All applications must use the same logging library and implementation to ensure consistency and ease of usage by humans and tools reading
the logs.
All applications must use the [logging](../../pkg/logging) package to log messages.

# TODO: verify logging implementation path.

By using the same library, we can ensure that all logs have the same format and structure.
Using one and only one library also improves the maintainability of the codebase, reduces the learning curve for new developers.
This reduces the number of dependencies in the codebase and reduces potential vulnerabilities.

### Initialization

The logger must be initialized at the beginning of the application execution.
By default, the logger must log messages at the `Info` level and above.
The `Debug` level must be enabled by a **debug** configuration parameter.
The parameter can be a command line flag, environment variable, or configuration file and must be searched in this order.
If the logging is implemented on a package level, the logger must be initialized as a no-op logger by default.

### Log Output

All logs must be stored in a Google Cloud Logging.
All logs written to the console must write to standard output for debug,
info, and warning logs, and standard error for errors and critical logs.
The logs written to the Google Cloud Logging must use proper marking an errors or critical logs, so they can be identified and properly
handled.

# TODO: Add example of logging error for golang application.

## Reference implementation