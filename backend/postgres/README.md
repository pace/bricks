# PostgreSQL


## Environment based configuration

Configuration for the PostgreSQL connection pool of the microservice.

* `POSTGRES_PORT` default: `5432`
    * Port to be used for listening used if address is not specified
* `POSTGRES_HOST` default: `localhost`
    * Host where the PostgreSQL can be found (dns or IP)
* `POSTGRES_PASSWORD` default: `pace1234!`
    * password to access the database
* `POSTGRES_USER` default: `postgres`
    * postgres user to access the database
* `POSTGRES_DB` default: `postgres`
    * database to access
* `POSTGRES_DIAL_TIMEOUT` default: `5s`
    * Dial timeout for establishing new connections
* `POSTGRES_READ_TIMEOUT` default: `30s`
    *  Timeout for socket reads. If reached, commands will fail with a timeout instead of blocking
* `POSTGRES_WRITE_TIMEOUT` default: `30s`
    * Timeout for socket writes. If reached, commands will fail with a timeout instead of blocking.
* `POSTGRES_HEALTH_CHECK_TABLE_NAME` default: `healthcheck`
    * Name of the Table that is created to try if database is writeable
* `POSTGRES_HEALTH_CHECK_RESULT_TTL` default: `10s`
    * Amount of time to cache the last health check result

## Metrics

Prometheus metrics exposed.

* `pace_postgres_query_total{database}` Collects stats about the number of postgres queries made
* `pace_postgres_query_failed{database}` Collects stats about the number of postgres queries failed
* `pace_postgres_query_duration_seconds{database}` Collects performance metrics for each postgres query
* `pace_postgres_query_affected_total{database}` Collects stats about the number of rows affected by a postgres query
* `pace_postgres_connection_pool_hits{database}` Collects number of times free connection was found in the pool
* `pace_postgres_connection_pool_misses{database}` Collects number of times free connection was NOT found in the pool
* `pace_postgres_connection_pool_timeouts{database}` Collects number of times a wait timeout occurred
* `pace_postgres_connection_pool_total_conns{database}` Collects number of total connections in the pool
* `pace_postgres_connection_pool_idle_conns{database}` Collects number of idle connections in the pool
* `pace_postgres_connection_pool_stale_conns{database}` Collects number of stale connections removed from the pool
