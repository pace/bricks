# Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
.PHONY: install test jsonapi build integration ci

JSONAPITEST=http/jsonapi/generator/internal
JSONAPIGEN="./tools/jsonapigen/main.go"
GOPATH?=~/go

GO:=go
GO_TEST_FLAGS:=-mod=vendor -count=1 -v -cover -race
PROTO_TMP:=$(shell pwd)/proto.tmp

export JAEGER_SERVICE_NAME:=unittest
export JAEGER_SAMPLER_TYPE:=const
export JAEGER_SAMPLER_PARAM:=1
export LOG_FORMAT:=console

install:
	$(GO) install ./cmd/pb

vuln-scan:
	$(GO) run -mod=vendor golang.org/x/vuln/cmd/govulncheck ./...

jsonapi:
	$(GO) run $(JSONAPIGEN) -pkg poi \
		-path $(JSONAPITEST)/poi/open-api_test.go \
		-source $(JSONAPITEST)/poi/open-api.json
	$(GO) run $(JSONAPIGEN) --pkg fueling \
		-path $(JSONAPITEST)/fueling/open-api_test.go \
		-source $(JSONAPITEST)/fueling/open-api.json
	$(GO) run $(JSONAPIGEN) -pkg pay \
		-path $(JSONAPITEST)/pay/open-api_test.go \
		-source $(JSONAPITEST)/pay/open-api.json
	$(GO) run $(JSONAPIGEN) -pkg articles \
		-path $(JSONAPITEST)/articles/open-api_test.go \
		-source $(JSONAPITEST)/articles/open-api.json
	$(GO) run $(JSONAPIGEN) -pkg securitytest \
		-path $(JSONAPITEST)/securitytest/open-api_test.go \
		-source $(JSONAPITEST)/securitytest/open-api.json
	$(GO) run $(JSONAPIGEN) -pkg simple \
		-path tools/testserver/simple/open-api.go \
		-source tools/testserver/simple/open-api.json

grpc: tools/testserver/math/math.pb.go

tools/testserver/math/math.pb.go: tools/testserver/math/math.proto
	mkdir -p $(PROTO_TMP)
	GOBIN=$(PROTO_TMP) $(GO) install -mod=vendor google.golang.org/grpc/cmd/protoc-gen-go-grpc
	GOBIN=$(PROTO_TMP) $(GO) install -mod=vendor google.golang.org/protobuf/cmd/protoc-gen-go
	protoc --plugin=$(PROTO_TMP)/protoc-gen-go-grpc \
		--plugin=$(PROTO_TMP)/protoc-gen-go \
		-I=./ --go-grpc_out=$(dir @) --go_out=$(dir @) $<
	rm -rf $(PROTO_TMP)

lint:
	$(GO) run -mod=vendor github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout 2m

test:
	$(GO) test $(GO_TEST_FLAGS) -covermode=atomic -coverprofile=coverage.out -short ./...

integration:
	$(GO) test $(GO_TEST_FLAGS) -run TestIntegration ./...
	$(GO) test $(GO_TEST_FLAGS) -run Example_clusterBackgroundTask ./pkg/routine

testserver:
	docker-compose up

ci:
	$(GO) test $(GO_TEST_FLAGS) -covermode=atomic -coverprofile=coverage.out ./...
