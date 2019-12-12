# Health Checks
* Makes it possible to add a health check for a service (e.g. postgres). The list of checks is checked for the routes
`/health` and `/health/check`

* `/health` all required and registered health checks healthy: Status code 200, Content-Type: text/plain, body: "OK" 
 
    Any required and registered health check is not healthy: 
 status code 503, Content-Type: text/plain, body: "ERR" 

* `/health/check` returns a table with the results of all registered health checks

Example:
``` 
    Required Services: 
    
    postgresdefault        ERR   any error message
    
    redis                  OK    
    
    Optional Services:  
    anotherTestName        WARN  any warning message

```

* Each check has to implement the `HealthCheck` interface and has to be registered
    * health checks are registered as required or optional 
    * `servicehealthcheck.RegisterHealthCheck(hc HealthCheck, name string)` registers a health check with an unique 
    name as required check
    * `servicehealthcheck.RegisterOptionalHealthCheck(hc HealthCheck, name string)` registers a health check with 
    a unique name as optional check
    * a optional check is not checked when `/health` is called
    * a required check is always checked
    
*  To change a registered health check from required to optional or vice versa the health check has to be removed and 
registered again. Remove a health check with `servicehealthcheck.RemoveHealthCheck(name)` 

## Implement a `HealthCheck`
* The result of each health check is not NOT cached. The implementation of each health check can use `ConnectionState` 
for caching 

* The `HealthCheck.HealthCheck()` returns OK, WARN or ERR and a detailed message
    * `/health` => OK and WARN means the service is healthy. 
    * `/health/check` => the complete result of the check is added to the response 
    
