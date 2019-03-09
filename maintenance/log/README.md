# Logging

Logging will be to STDOUT, then captured by the fluentd and finally delivered to ElasticSearch or Graylog.

**Reasoning:**

* Local log storage allows for easy maintenance on ELK/GL, no problem on outage for production system

#### Format

The format is implemented in the log package and generally refered to as KVP stile logging.
It is easy to parse and generate in code when logging to STDOUT and not as heavy as JSON on the eyes. See example below:

```
2018-09-06 15:31:23 |INFO| Starting testserver ... addr=:3000
```

The KVP logs are only generated if the stdout device is a terminal. If the device is not a terminal
JSON output will be generated. This way there is very little parse overhead for imports into storage
systems like elasticsearch.

**JSON log format attributes**

|Name|Type|Example|Notes|
|-|-|-|-|
| level | `string` | `"info"` |
| req_id | `string` | `"be922tboo3smmkmppjfg"` | The generated id is a URL safe base64 <br/>encoded mongo object-id-like unique id.<br/> Mongo unique id generation algorithm has<br/> been selected as a trade-off between size<br/> and ease of use: UUID is less space efficient<br/> and snowflake re`quires machine configuration. |
| method | `string` | `"GET"` |
| url | `string` | `"/foo/bar"` |
| status | `int` | `501` |
| host | `string` | `"example.com"` | taken from the request |
| size | `int` | `121` |
| duration | `float` | `0.217762` |
| ip | `string` | `"192.0.2.1"` | respects `X-Forwarded-For` and `X-Real-Ip` |
| referer | `string` | `"https://google.de"` |
| user_agent | `string` | `"Mozilla/5.0 (Macintosh;"` |
| time | `string` | `"2018-09-07 06:57:57"` | iso8601 UTC |
| message | `string` | `"Request Completed"` |
|-|-| **Microservice specific** |-|
| handler | `string` | `"GetPumpHandler"` | Name of the handler func in case of a panic |
| error | `string` | `"Can't open file"` | text representation of the error |

**Constraints**
* Encoding of all entries is in `UTF-8`
* Timestamps are `UTC` format is `iso8601`
* Use keys consistently

## Environment based configuration

* `LOG_LEVEL` default: `debug`
    * Allows for logging at the following levels (from highest to lowest): 
        * `panic`
        * `fatal`
        * `error`
        * `warn`
        * `info`
        * `debug`
        * `disabled` don't log at all
* `LOG_FORMAT` default: `auto`
    * If set to auto will detect if stdout is attached to a TTY and set the format to `console`
      otherwise the format will be `json`. Formats can be set directly.

## Resources

* https://logz.io/blog/logging-best-practices/
* https://docs.logentries.com/docs/best-practices-logs
