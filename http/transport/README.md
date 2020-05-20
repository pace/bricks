# HTTP Transports

HTTP Transport to be used with the http client from go.

## Environment based configuration

* `HTTP_TRANSPORT_DUMP` default: `""` (option values are comma separated)
  * Can contain a list of logging options:
    * `request` will log the complete request with headers human readable
    * `response` will log the complete response with headers human readable
    * `request-hex` will log the complete request with headers in HEX
    * `response-hex` will log the complete response with headers in HEX
    * `body` will enable logging of the body for hex and human readable outputs
  * Simple request header logging may look like this `request,response`
  * Full human readable logging may look like this `request,response,body`
  * Complete logging may look like this `request,response,request-hex,response-hex,body`
