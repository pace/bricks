# HTTP server

## Environment based configuration

* `ADDR`
    * Address golang listen address in the [Dial format](https://golang.org/pkg/net/#Dial)
* `PORT` default: `3000`
    * Port to be used for listening used if address is not specified
* `ENVIRONMENT` default: `edge`
    * Name of the current environment
* `MAX_HEADER_BYTES` default: `1048576` (1 MB)
    * MaxHeaderBytes controls the maximum number of bytes the
      server will read parsing the request header's keys and
      values, including the request line. It does not limit the
      size of the request body. If zero, DefaultMaxHeaderBytes is used.
* `IDLE_TIMEOUT` default: `1h`
    * IdleTimeout is the maximum amount of time to wait for the
      next request when keep-alives are enabled. If IdleTimeout
      is zero, the value of ReadTimeout is used. If both are
      zero, ReadHeaderTimeout is used.
    * Everything that can be parsed by [ParseDuration](https://golang.org/pkg/time/#ParseDuration)
* `READ_TIMEOUT` default: `60s`
    * ReadHeaderTimeout is the amount of time allowed to read
      request headers. The connection's read deadline is reset
      after reading the headers and the Handler can decide what
      is considered too slow for the body.
    * Everything that can be parsed by [ParseDuration](https://golang.org/pkg/time/#ParseDuration)
* `WRITE_TIMEOUT` default: `60s`
    * WriteTimeout is the maximum duration before timing out
      writes of the response. It is reset whenever a new
      request's header is read. Like ReadTimeout, it does not
      let Handlers make decisions on a per-request basis.
    * Everything that can be parsed by [ParseDuration](https://golang.org/pkg/time/#ParseDuration)