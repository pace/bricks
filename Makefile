# Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
# Created at 2018/08/24 by Vincent Landgraf

JSONAPITEST="http/jsonapi/generator/internal"
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
