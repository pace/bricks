module github.com/pace/bricks

go 1.16

replace github.com/adjust/rmq/v3 => github.com/daemonfire300/rmq/v3 v3.0.2

require (
	github.com/adjust/rmq/v3 v3.0.0
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/bsm/redislock v0.5.0
	github.com/caarlos0/env v3.3.0+incompatible
	github.com/certifi/gocertifi v0.0.0-20180118203423-deb3ae2ef261
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/dave/jennifer v1.0.2
	github.com/getkin/kin-openapi v0.0.0-20180813063848-e1956e8013e5
	github.com/go-kivik/couchdb/v3 v3.2.6
	github.com/go-kivik/kivik/v3 v3.2.3
	github.com/go-pg/pg v6.14.5+incompatible
	github.com/go-redis/redis/v7 v7.4.1
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/golangci/golangci-lint v1.42.0
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/jpillora/backoff v1.0.0
	github.com/mattn/go-isatty v0.0.12
	github.com/mattn/goveralls v0.0.9
	github.com/minio/minio-go/v7 v7.0.7
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/prometheus/client_golang v1.7.1
	github.com/rs/xid v1.2.1
	github.com/rs/zerolog v1.17.2
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114
	github.com/sony/gobreaker v0.4.1
	github.com/spf13/cobra v1.2.1
	github.com/streadway/handy v0.0.0-20200128134331-0f66f006fb2e
	github.com/stretchr/testify v1.7.0
	github.com/uber-go/atomic v1.3.2 // indirect
	github.com/uber/jaeger-client-go v2.14.0+incompatible
	github.com/uber/jaeger-lib v1.5.0
	github.com/zenazn/goji v0.9.0
	golang.org/x/mod v0.5.0 // indirect
	golang.org/x/tools v0.1.5
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/genproto v0.0.0-20210719143636-1d5a45f8e492 // indirect
	google.golang.org/grpc v1.39.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.27.1
)

replace github.com/codahale/hdrhistogram => github.com/HdrHistogram/hdrhistogram-go v0.9.0
