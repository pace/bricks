FROM golang:1.12 as builder
WORKDIR /tmp/pace-bricks
ADD . .

# Build go files completely statically
RUN GOPATH=/tmp/go CGO_ENABLED=0 go build -mod=vendor -a -ldflags '-extldflags "-static"' -o /usr/local/bin/pb-testserver ./tools/testserver
RUN GOPATH=/tmp/go CGO_ENABLED=0 go build -mod=vendor -a -ldflags '-extldflags "-static"' -o /usr/local/bin/pb ./cmd/pb

EXPOSE 5000

ENV PORT 5000
ENV JAEGER_SERVICE_NAME pace-bricks
ENV LOG_FORMAT console

CMD /usr/local/bin/pb-testserver
