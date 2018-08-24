# Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
# Created at 2018/08/24 by Vincent Landgraf

GOPATH:="$(HOME)/go"
PACETOOL_PKG="lab.jamit.de/pace/web/tool/cmd/pace"
PACE="$(GOPATH)/src/$(PACETOOL_PKG)/pace.go"

JSONAPITEST="http/jsonapi/generator/internal"

jsonapi:
	go run $(PACE) service generate rest --pkg poi \
		--path $(JSONAPITEST)/poi/open-api_test.go \
		--source $(JSONAPITEST)/poi/open-api.json
	go run $(PACE) service generate rest --pkg fueling \
		--path $(JSONAPITEST)/fueling/open-api_test.go \
		--source $(JSONAPITEST)/fueling/open-api.json
	go run $(PACE) service generate rest --pkg pay \
		--path $(JSONAPITEST)/pay/open-api_test.go \
		--source $(JSONAPITEST)/pay/open-api.json
