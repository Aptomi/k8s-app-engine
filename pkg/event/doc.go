// Package event implements support for Aptomi Event Logs and saving them to console, memory, and external
// stores (BoltDB).
// Event logs are user-friendly logs (e.g. policy resolution log, policy apply log), which eventually get
// shown to the end-users through UI. Unlike standard logs, event logs are fully stored in memory before
// they get persisted. This is required in order for the engine to attach "details" to every log record.
// This package also provides a mock logger, which can be useful in unit tests.
package event
