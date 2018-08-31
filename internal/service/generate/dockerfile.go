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
RUN go get gopkg.in/alecthomas/gometalinter.v2
RUN gometalinter.v2 --install
WORKDIR /tmp/service
ADD . .
RUN gometalinter.v2 ./...
RUN go test -v -race -cover ./...
RUN go install ./cmd/{{ .Commands.DaemonName }}
RUN go install ./cmd/{{ .Commands.ControlName }}

FROM alpine
RUN apk update && apk add ca-certificates && apk add tzdata && rm -rf /var/cache/apk/*
COPY --from=builder /go/bin/{{ .Commands.DaemonName }} /usr/local/bin/
COPY --from=builder /go/bin/{{ .Commands.ControlName }} /usr/local/bin/

EXPOSE 3000
ENV PORT 3000
ENTRYPOINT ["/usr/local/bin/{{ .Commands.DaemonName }}"]
`))
