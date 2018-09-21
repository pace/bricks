# Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
# Created at 2018/08/24 by Vincent Landgraf
.PHONY: install test jsonapi docker.all docker.jaeger docker.postgres docker.postgres.setup docker.redis

JSONAPITEST=http/jsonapi/generator/internal
JSONAPIGEN="./tools/jsonapigen/main.go"
PGPASSWORD="pace1234!"
psql=PGPASSWORD="pace1234!" psql -h localhost -U postgres

install:
	go install ./cmd/pace

jsonapi:
	go run $(JSONAPIGEN) -pkg poi \
		-path $(JSONAPITEST)/poi/open-api_test.go \
		-source $(JSONAPITEST)/poi/open-api.json
	go run $(JSONAPIGEN) --pkg fueling \
		-path $(JSONAPITEST)/fueling/open-api_test.go \
		-source $(JSONAPITEST)/fueling/open-api.json
	go run $(JSONAPIGEN) -pkg pay \
		-path $(JSONAPITEST)/pay/open-api_test.go \
		-source $(JSONAPITEST)/pay/open-api.json

test:
	JAEGER_SERVICE_NAME=unittest LOG_FORMAT=console go test -v -race -cover ./...

testserver:
	JAEGER_ENDPOINT=http://localhost:14268/api/traces \
	JAEGER_SAMPLER_TYPE=const \
	JAEGER_SAMPLER_PARAM=1 \
	JAEGER_SERVICE_NAME=testserver \
	SENTRY_DSN="https://71e5037808ff4acf9e18cd0ab5ee472a:f58b0c9ba48447af953cf377cf1d9b9c@sentry.jamit.de/164" \
	SENTRY_ENVIRONMENT=development \
	SENTRY_RELEASE=`git rev-parse --short HEAD` \
	POSTGRES_USER=testserveruser \
	POSTGRES_DB=testserver \
	POSTGRES_PASSWORD=pace1234! \
	go run ./tools/testserver/main.go

docker.all: docker.jaeger docker.postgres.setup docker.redis

docker.jaeger:
	docker run -d --rm --name jaeger \
		-e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
		-p 5775:5775/udp \
		-p 6831:6831/udp \
		-p 6832:6832/udp \
		-p 5778:5778 \
		-p 16686:16686 \
		-p 14268:14268 \
		-p 9411:9411 \
		jaegertracing/all-in-one:latest

docker.postgres:
	docker run -d --rm --name postgres \
		-e POSTGRES_PASSWORD=$(PGPASSWORD) \
		-p 5432:5432 \
		postgres:9

docker.postgres.setup: docker.postgres
	$(psql) -c "CREATE DATABASE testserver;"
	$(psql) -c "CREATE USER testserveruser WITH ENCRYPTED PASSWORD 'pace1234!';"
	$(psql) -c "GRANT ALL PRIVILEGES ON DATABASE testserver TO testserveruser;"

docker.redis:
	docker run -d --rm --name redis \
		-p 6379:6379 \
		redis
