# Health Checks for Services 
* Makes it possible to have a health check for different services (e.g. postgres, redis)

* A service has to implement the HealthCheck interface and register the health check with `healthchecker.RegisterHealthCheck(hc HealthCheck)`

* `healthchecker` does NOT cache the results but offers a struct (`ConnectionState`) for caching 

