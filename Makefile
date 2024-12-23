# Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
JSONAPITEST=http/jsonapi/generator/internal
JSONAPIGEN="./tools/jsonapigen/main.go"
GOPATH?=~/go

GO:=go
GO_TEST_FLAGS:=-count=1 -v -cover -race

export JAEGER_SERVICE_NAME:=unittest
export LOG_FORMAT:=console

.PHONY: install
install:
	$(GO) install ./cmd/pb

.PHONY: vuln-scan
vuln-scan:
	(cd /; $(GO) install -v -x golang.org/x/vuln/cmd/govulncheck@latest)

	govulncheck ./...

.PHONY: jsonapi
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

.PHONY: grpc
grpc: tools/testserver/math/math.pb.go

tools/testserver/math/math.pb.go: tools/testserver/math/math.proto
	(cd /; $(GO) install -v -x google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest)
	(cd /; $(GO) install -v -x google.golang.org/protobuf/cmd/protoc-gen-go@latest)

	protoc -I=./ --go-grpc_out=$(dir @) --go_out=$(dir @) $<

.PHONY: lint
lint:
	(cd /; $(GO) install -x github.com/golangci/golangci-lint/cmd/golangci-lint@latest)

	golangci-lint run --timeout 2m

.PHONY: test
test:
	$(GO) test $(GO_TEST_FLAGS) -covermode=atomic -coverprofile=coverage.out -short ./...

.PHONY: integration
integration:
	$(GO) test $(GO_TEST_FLAGS) -run TestIntegration ./...
	$(GO) test $(GO_TEST_FLAGS) -run Example_clusterBackgroundTask ./pkg/routine

.PHONY: testserver
testserver:
	docker-compose up

.PHONY: ci
ci:
	$(GO) test $(GO_TEST_FLAGS) -covermode=atomic -coverprofile=coverage.out ./...
