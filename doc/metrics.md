## Metrics

To provide metrics all services have to implement the Prometheus metrics API.

The defined metrics should follow the best practices defined [here](https://prometheus.io/docs/practices/naming/).

### Go VM Metrics

* `go_*`
    * all microservice will include the [GoCollector](https://github.com/prometheus/client_golang/blob/master/prometheus/go_collector.go) provided metrics
    * Use cases:
        * Track goroutines/theads/GC for developers and DevOps
        * Trank system uage (saturation light) for DevOps (**Golden Signal**)

	* `go_goroutines`
		* Number of goroutines that currently exist
	* `go_threads`
		* Number of OS threads created
	* `go_gc_duration_seconds`
		* A summary of the GC invocation durations
	* `go_info`
		* Information about the Go environment
        * Labels:
            * **version** (go1.10.3, ....)
    * `go_memstats_alloc_bytes`
        * Number of bytes allocated and still in use
    * `go_memstats_alloc_bytes_total`
        * Total number of bytes allocated, even if freed
    * `go_memstats_sys_bytes`
        * Number of bytes obtained from system
    * `go_memstats_lookups_total`
        * Total number of pointer lookups
    * `go_memstats_mallocs_total`
        * Total number of mallocs
    * `go_memstats_frees_total`
        * Total number of frees
    * `go_memstats_heap_alloc_bytes`
        * Number of heap bytes allocated and still in use
    * `go_memstats_heap_sys_bytes`
        * Number of heap bytes obtained from system
    * `go_memstats_heap_idle_bytes`
        * Number of heap bytes waiting to be used
    * `go_memstats_heap_inuse_bytes`
        * Number of heap bytes that are in use
    * `go_memstats_heap_released_bytes`
        * Number of heap bytes released to OS
    * `go_memstats_heap_objects`
        * Number of allocated objects
    * `go_memstats_stack_inuse_bytes`
        * Number of bytes in use by the stack allocator
    * `go_memstats_stack_sys_bytes`
        * Number of bytes obtained from system for stack allocator
    * `go_memstats_mspan_inuse_bytes`
        * Number of bytes in use by mspan structures
    * `go_memstats_mspan_sys_bytes`
        * Number of bytes used for mspan structures obtained from system
    * `go_memstats_mcache_inuse_bytes`
        * Number of bytes in use by mcache structures
    * `go_memstats_mcache_sys_bytes`
        * Number of bytes used for mcache structures obtained from system
    * `go_memstats_buck_hash_sys_bytes`
        * Number of bytes used by the profiling bucket hash table
    * `go_memstats_gc_sys_bytes`
        * Number of bytes used for garbage collection system metadata
    * `go_memstats_other_sys_bytes`
        * Number of bytes used for other system allocations
    * `go_memstats_next_gc_bytes`
        * Number of heap bytes when next garbage collection will take place
    * `go_memstats_last_gc_time_seconds`
        * Number of seconds since 1970 of last garbage collection
    * `go_memstats_gc_cpu_fraction`
        * The fraction of this program's available CPU time used by the GC since the program started

## Process Metrics

* `process_*` 
    * provided by the [ProcessCollector](https://github.com/prometheus/client_golang/blob/master/prometheus/process_collector.go#L34)
    * `process_cpu_seconds_total`
        * Total user and system CPU time spent in seconds
    * `process_open_fds`
    	* Number of open file descriptors
    * `process_max_fds`
        * Maximum number of open file descriptors
    * `process_virtual_memory_bytes`
		* Virtual memory size in bytes
	* `process_virtual_memory_max_bytes`
		* Maximum amount of virtual memory available in bytes
    * `process_resident_memory_bytes`
		* Resident memory size in bytes
	* `process_start_time_seconds`
		* Start time of the process since unix epoch in seconds

## Prometheus metrics

* `promhttp_*`
    * Prometheus related counters documented [here](https://github.com/prometheus/client_golang/blob/master/prometheus/promhttp/http.go), provided by the default handler
    * `promhttp_metric_handler_requests_total`
        * Total number of scrapes by HTTP status code.
        * Labels:
            * **Code** (200, 501, ...) - HTTP status code
    * `promhttp_metric_handler_requests_in_flight`
        * Current number of scrapes being served. This function idempotently registers collectors for both metrics with the provided Registerer. It panics if the registration fails. The provided metrics are useful to see how many scrapes hit the monitored target (which could be from different Prometheus servers or other  scrapers), and how often they overlap (which would result in more than one scrape in flight at the same time). Note that the scrapes-in-flight gauge will contain the scrape by which it is exposed, while the scrape counter will only get incremented after the scrape is complete (as only then the status code is known). For tracking scrape durations, use the "scrape_duration_seconds" gauge created by the Prometheus server upon each scrape.

## Pace Bricks specific metrics

### HTTP Request Metrics 

* `pace_api_http_request_total` (Counter)
    * Collects statistics about each microservice endpoint
    * Use cases:
        * Track API usage for business partners (ClientID)
        * Track error rate per service path/method (**Golden Signal**)
        * Track API usage for developers (Path and Method) for deprecations
        * Track request rate (**Golden Signal**)
    * Labels:
        * **Code** (200, 501, ...) - HTTP status code
        * **Method** (GET, PUT, POST, ...) - HTTP method
        * **Path** ("/beta/cars", "/beta/cars/{id}", ...) - Path to the endpoint as defined in the OpenAPIv3 spec
        * **Service** (car, dtc, ...) - name of the microservice
        * **ClientID** (unknown, "XYZ") - OAuth2 ClientID to filter usage per business partner, app, cockpit, ...

* `pace_api_http_request_duration_seconds` (Histogram)
    * Collect performance metrics for each API endpoint
    * Use cases:
        * Track latency per service path/method (**Golden Signal**)
        * Identify low performing service endpoints for development / identify performance regressions
    * Labels:
        * **Method** (GET, PUT, POST, ...) - HTTP method
        * **Path** ("/beta/cars", "/beta/cars/{id}", ...) - Path to the endpoint as defined in the OpenAPIv3 spec
        * **Service** (car, dtc, ...) - name of the microservice

* `pace_api_http_size_bytes` (Histogram)
    * Collect performance metrics for each API endpoint
    * Use cases:
        * Track amount of bytes responded/requested per service path/method (**Golden Signal**)
        * Identify possible issues for the application due to to big responses
    * Labels:
        * **Method** (GET, PUT, POST, ...) - HTTP method
        * **Path** ("/beta/cars", "/beta/cars/{id}", ...) - Path to the endpoint as defined in the OpenAPIv3 spec
        * **Service** (car, dtc, ...) - name of the microservice
        * **Type** (req, resp) - HTTP request or response
