# Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
# Created at 2018/08/24 by Vincent Landgraf
.PHONY: install test jsonapi build integration

JSONAPITEST=http/jsonapi/generator/internal
JSONAPIGEN="./tools/jsonapigen/main.go"

export JAEGER_SERVICE_NAME:=unittest
export JAEGER_SAMPLER_TYPE:=const
export JAEGER_SAMPLER_PARAM:=1
export LOG_FORMAT:=console

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

build:
	docker build .

lint:
	gometalinter --cyclo-over=15 --deadline 90s --skip http/jsonapi/generator/internal --skip tools --vendor ./...

test:
	go test -count=1 -v -cover -race -short ./...

integration: build
	go test -count=1 -v -cover -race -run TestIntegration ./...

testserver:
	docker build .
	docker-compose up
