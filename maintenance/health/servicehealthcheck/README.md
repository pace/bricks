# Health Checks for Services 
* makes it possible to have a health check for different services (e.g. postgres, redis)
* A service has to implement the HealthCheck interface and register the healthCheck with `healthchecker.RegisterHealthCheck(hc HealthCheck)`
* `healthchecker` does NOT make any caching of the results