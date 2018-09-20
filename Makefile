# Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
# Created at 2018/08/24 by Vincent Landgraf
.PHONY: install jsonapi docker.all docker.jaeger docker.postgres docker.redis

JSONAPITEST=http/jsonapi/generator/internal
JSONAPIGEN="./tools/jsonapigen/main.go"

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

testserver:
	JAEGER_ENDPOINT=http://localhost:14268/api/traces \
	JAEGER_SAMPLER_TYPE=const \
	JAEGER_SAMPLER_PARAM=1 \
	JAEGER_SERVICE_NAME=testserver \
	POSTGRES_USER=testserveruser \
	POSTGRES_DB=testserver \
	POSTGRES_PASSWORD=pace1234! \
	go run ./tools/testserver/main.go

docker.all: docker.jaeger docker.postgres docker.redis

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
		-e POSTGRES_PASSWORD=pace1234! \
		-p 5432:5432 \
		postgres:9

docker.redis:
	docker run -d --rm --name redis \
		-p 6379:6379 \
		redis