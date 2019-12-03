# Health Check
* Makes it possible to add a health check for a service (e.g. postgres) to the list of checks checked when for `/health` and `/health/check`

* `/health` returns OK(200) if all registered and required health checks are healthy and ERR(503) if any health check is not healthy

* `/health/check` returns a table with the results of all registered health checks

* Each check has to implement the `HealthCheck` interface and has to be registered
    * health checks are registered as required or optional 
    * `servicehealthcheck.RegisterHealthCheck(hc HealthCheck, name string)` registers a health check with a unique name as required check
    * `servicehealthcheck.RegisterOptionalHealthCheck(hc HealthCheck, name string)` registers a health check with a unique name as optional check
    * a optional check is not checked when `/health` is called
    * a required check is always checked
    
* To make a optional health check required (or a required health check optional) the check has to be removed and registered again. This should't happen very often.
Remove a health check with `servicehealthcheck.RemoveHealthCheck(name)` 

## Implement a `HealthCheck`
* The result of each health check is not NOT cached. The implementation of each health check can use `ConnectionState` for caching 

* The `HealthCheck.HealthCheck()` returns OK, WARN or ERR and a detailed message
    * `/health` => OK and WARN means the service is healthy. 
    * `/health/check` => the complete result of the check is added to the response 
    
