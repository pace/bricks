# Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
# Created at 2018/08/24 by Vincent Landgraf
.PHONY: install test jsonapi build integration ci

JSONAPITEST=http/jsonapi/generator/internal
JSONAPIGEN="./tools/jsonapigen/main.go"
GOPATH?=~/go

export JAEGER_SERVICE_NAME:=unittest
export JAEGER_SAMPLER_TYPE:=const
export JAEGER_SAMPLER_PARAM:=1
export LOG_FORMAT:=console

install:
	go install ./cmd/pb

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
	go run $(JSONAPIGEN) -pkg articles \
		-path $(JSONAPITEST)/articles/open-api_test.go \
		-source $(JSONAPITEST)/articles/open-api.json

lint: $(GOPATH)/bin/golangci-lint
	$(GOPATH)/bin/golangci-lint run

$(GOPATH)/bin/golangci-lint:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(GOPATH)/bin v1.21.0

test:
	go test -mod=vendor -count=1 -v -cover -race -short -covermode=count -coverprofile=profile.cov ./...

integration:
	go test -mod=vendor -count=1 -v -cover -race -run TestIntegration -covermode=count -coverprofile=profile.cov ./...

testserver:
	docker-compose up

ci: test integration
