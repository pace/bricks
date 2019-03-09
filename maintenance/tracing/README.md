# Tracing (Jaeger)

All microservice will be using OpenTracing (with Jaeger via UDP).

## Environment based configuration

Configuration directly taken from https://github.com/jaegertracing/jaeger-client-go.

Property| Description
--- | ---
`JAEGER_SERVICE_NAME` | The service name
`JAEGER_AGENT_HOST` | The hostname for communicating with agent via UDP
`JAEGER_AGENT_PORT` | The port for communicating with agent via UDP
`JAEGER_ENDPOINT` | The HTTP endpoint for sending spans directly <br/>to a collector, i.e. `http://jaeger-collector:14268/api/traces`
`JAEGER_USER` | Username to send as part of "Basic" authentication<br/> to the collector endpoint
`JAEGER_PASSWORD` | Password to send as part of "Basic" authentication<br/> to the collector endpoint
`JAEGER_REPORTER_LOG_SPANS` | Whether the reporter should also log the spans
`JAEGER_REPORTER_MAX_QUEUE_SIZE` | The reporter's maximum queue size
`JAEGER_REPORTER_FLUSH_INTERVAL` | The reporter's flush interval (ms)
`JAEGER_SAMPLER_TYPE` | The sampler type
`JAEGER_SAMPLER_PARAM` | The sampler parameter (number)
`JAEGER_SAMPLER_MANAGER_HOST_PORT` | The HTTP endpoint when using the remote sampler,<br/> i.e. `http://jaeger-agent:5778/sampling`
`JAEGER_SAMPLER_MAX_OPERATIONS` | The maximum number of operations that the sampler<br/> will keep track of
`JAEGER_SAMPLER_REFRESH_INTERVAL` | How often the remotely controlled sampler will poll<br/> jaeger-agent for `the` appropriate sampling strategy
`JAEGER_TAGS` | A comma separated list of `name = value` tracer <br/>level tags, which get added to all `reported` spans.<br/> The value can also refer to an environment variable<br/> using the format `${envVarName:default}`, where<br/> the `:default` is optional, and identifies a value to be<br/> used if the environment variable cannot be found
`JAEGER_DISABLED` | Whether the tracer is disabled or not. If true, the default `opentracing.NoopTracer` is used.
`JAEGER_RPC_METRICS` | Whether to store RPC metrics
