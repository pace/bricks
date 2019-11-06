
## Environment based configuration

* `REDIS_HOSTS` default: `redis:6379`
    * host:port addresses, can be multiple separated by comma.
* `REDIS_PASSWORD`
    * Optional password. Must match the password specified in the `requirepass` server configuration option.
* `REDIS_DB`
    * Database to be selected after connecting to the server.
* `REDIS_MAX_RETRIES`
    * Maximum number of retries before giving up. Default is to not retry failed commands.
* `REDIS_POOL_SIZE` default: `10`
    * Maximum number of socket connections. Default is 10 connections per every CPU as reported by runtime.NumCPU.
* `REDIS_MIN_IDLE_CONNS`
    * Minimum number of idle connections which is useful when establishing new connection is slow.
* `REDIS_MIN_RETRY_BACKOFF` default: `8ms`
    * Minimum backoff between each retry. Default is 8 milliseconds; -1 disables backoff.
    * Everything that can be parsed by [ParseDuration](https://golang.org/pkg/time/#ParseDuration)
* `REDIS_MAX_RETRY_BACKOFF` default: `512ms`
    * Maximum backoff between each retry. Default is 512 milliseconds; -1 disables backoff.
    * Everything that can be parsed by [ParseDuration](https://golang.org/pkg/time/#ParseDuration)
* `REDIS_DIAL_TIMEOUT` default: `5s`
    * Dial timeout for establishing new connections. Default is 5 seconds.
    * Everything that can be parsed by [ParseDuration](https://golang.org/pkg/time/#ParseDuration)
* `REDIS_READ_TIMEOUT` default: `3s`
    * Timeout for socket reads. If reached, commands will fail with a timeout instead of blocking. Default is 3 seconds.
    * Everything that can be parsed by [ParseDuration](https://golang.org/pkg/time/#ParseDuration)
* `REDIS_WRITE_TIMEOUT`
    * Timeout for socket writes. If reached, commands will fail with a timeout instead of blocking. Default is ReadTimeout.
    * Everything that can be parsed by [ParseDuration](https://golang.org/pkg/time/#ParseDuration)
* `REDIS_MAX_CONNAGE`
    * Connection age at which client retires (closes) the connection. Default is to not close aged connections.
    * Everything that can be parsed by [ParseDuration](https://golang.org/pkg/time/#ParseDuration)
* `REDIS_POOL_TIMEOUT` default ReadTimeout + 1 second
    * Amount of time client waits for connection if all connections are busy before returning an error.
    * Everything that can be parsed by [ParseDuration](https://golang.org/pkg/time/#ParseDuration)
* `REDIS_IDLE_TIMEOUT` default: `5m`
    * Amount of time after which client closes idle connections. Should be less than server's timeout. Default is 5 minutes. -1 disables idle timeout check.
    * Everything that can be parsed by [ParseDuration](https://golang.org/pkg/time/#ParseDuration)
* `REDIS_IDLE_CHECK_FREQUENCY` default: `1m`
    * Frequency of idle checks made by idle connections reaper. Default is 1 minute. -1 disables idle connections reaper, but idle connections are still discarded by the client if IdleTimeout is set.
    * Everything that can be parsed by [ParseDuration](https://golang.org/pkg/time/#ParseDuration)
* `REDIS_HEALTH_KEY` default: `healthy`
    * Name of the key that is written to check, if redis is healthy
* `REDIS_HEALTHCHECK_MAX_REQUEST_SEC` default: `10s`
    * Amount of time to cache the last health check result
* `REDIS_HEALTHCHECK_WRITE` default: `true`
	* Whether also to perform a write test