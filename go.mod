module github.com/pace/bricks

go 1.16

replace github.com/adjust/rmq/v3 => github.com/daemonfire300/rmq/v3 v3.0.2

require (
	github.com/adjust/rmq/v3 v3.0.0
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973 // indirect
	github.com/bsm/redislock v0.5.0
	github.com/caarlos0/env v3.3.0+incompatible
	github.com/certifi/gocertifi v0.0.0-20180118203423-deb3ae2ef261
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/dave/jennifer v1.0.2
	github.com/getkin/kin-openapi v0.0.0-20180813063848-e1956e8013e5
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-kivik/couchdb/v3 v3.2.6
	github.com/go-kivik/kivik/v3 v3.2.3
	github.com/go-pg/pg v6.14.5+incompatible
	github.com/go-redis/redis/v7 v7.4.1
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/mux v1.6.2
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/jpillora/backoff v1.0.0
	github.com/mattn/go-isatty v0.0.11
	github.com/mattn/goveralls v0.0.5
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/minio/minio-go/v7 v7.0.7
	github.com/opentracing/opentracing-go v1.0.2
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/prometheus/client_golang v0.8.0
	github.com/prometheus/client_model v0.0.0-20180712105110-5c3871d89910 // indirect
	github.com/prometheus/common v0.0.0-20180801064454-c7de2306084e // indirect
	github.com/prometheus/procfs v0.0.0-20180725123919-05ee40e3a273 // indirect
	github.com/rs/xid v1.2.1
	github.com/rs/zerolog v1.17.2
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114
	github.com/sony/gobreaker v0.4.1
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.2 // indirect
	github.com/streadway/handy v0.0.0-20200128134331-0f66f006fb2e
	github.com/stretchr/testify v1.6.1
	github.com/uber-go/atomic v1.3.2 // indirect
	github.com/uber/jaeger-client-go v2.14.0+incompatible
	github.com/uber/jaeger-lib v1.5.0
	github.com/zenazn/goji v0.9.0
	golang.org/x/tools v0.0.0-20200304024140-c4206d458c3f
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace github.com/codahale/hdrhistogram => github.com/HdrHistogram/hdrhistogram-go v0.9.0
