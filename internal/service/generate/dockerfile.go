// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/31 by Vincent Landgraf

package generate

import (
	"html/template"
	"log"
	"os"
)

// DockerfileOptions configure the output of the generated docker
// file
type DockerfileOptions struct {
	Name     string
	Commands CommandOptions
}

// Dockerfile generate a dockerfile using the given options
// for specified path
func Dockerfile(path string, options DockerfileOptions) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}

	err = dockerTemplate.Execute(f, options)
	if err != nil {
		log.Fatal(err)
	}
}

var dockerTemplate = template.Must(template.New("Dockerfile").Parse(
	`FROM golang:1.11 as builder
RUN go get github.com/alecthomas/gometalinter
RUN gometalinter --install
WORKDIR /tmp/service
ADD . .

# Lin, vet & test
# (many linters from gometalinter don't support go mod and therefore need to be disabled)
RUN gometalinter --disable-all --vendor -E gocyclo -E goconst -E golint -E ineffassign -E gotypex -E deadcode ./... && \
	go vet -mod vendor ./... && \
	go test -mod vendor -v -race -cover ./...

# Build go files completely statically
RUN CGO_ENABLED=0 go build -mod vendor  -a -ldflags '-extldflags "-static"' -o $GOPATH/bin/{{ .Commands.DaemonName }} ./cmd/{{ .Commands.DaemonName }} && \
	CGO_ENABLED=0 go build -mod vendor  -a -ldflags '-extldflags "-static"' -o $GOPATH/bin/{{ .Commands.ControlName }} ./cmd/{{ .Commands.ControlName }}

FROM alpine
RUN apk update && apk add ca-certificates && apk add tzdata && rm -rf /var/cache/apk/*
COPY --from=builder /go/bin/{{ .Commands.DaemonName }} /usr/local/bin/
COPY --from=builder /go/bin/{{ .Commands.ControlName }} /usr/local/bin/

EXPOSE 3000
ENV PORT 3000
ENV JAEGER_SERVICE_NAME {{ .Name }}
CMD /usr/local/bin/{{ .Commands.DaemonName }}
`))
