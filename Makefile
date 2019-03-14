# Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
# Created at 2018/08/24 by Vincent Landgraf
.PHONY: install test jsonapi

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

lint:
	gometalinter --cyclo-over=15 --deadline 90s --skip http/jsonapi/generator/internal --skip tools --vendor ./...

test:
	JAEGER_SERVICE_NAME=unittest JAEGER_SAMPLER_TYPE=const JAEGER_SAMPLER_PARAM=1 LOG_FORMAT=console go test -v -race -cover ./...

testserver:
	docker build .
	docker-compose up
