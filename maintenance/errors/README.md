# Error Handling (sentry)

* Each microservice need to expose `"/health"` and return with http 200, this is important for the load balancing of the API gateway
* Unhandled exception handling ("in the wild")
* Frontend Application 
* Returning 5xx to client (can indicate operations issues)
* Depending on application logic (additional errors; cases that should not happen; security)
* Implausible responses from upstream services (e.g. after validation of result from external API)

## Environment based configuration

* `SENTRY_DSN`
    * URL of the sentry DSN
* `SENTRY_ENVIRONMENT`
    * Environment of sentry reported in the dashboard
* `SENTRY_RELEASE`
    * Name of the release e.g. git commit or similar