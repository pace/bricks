# Tracing (Sentry)

All microservice will be using Sentry.

## Environment based configuration

Property| Description
--- | ---
`SENTRY_DSN`  | The DSN to use.
`ENVIRONMENT` | The environment to be sent with events.
`SENTRY_TRACES_SAMPLE_RATE` | The tracing sample rate to use (default: 0.1).
`SENTRY_ENABLE_TRACING` | Enable or disable tracing (default: true).
