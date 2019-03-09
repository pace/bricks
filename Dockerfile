######### BUILDER

FROM golang:1.12 as builder
WORKDIR /tmp/pace-bricks
ADD . .

# Build go files completely statically
RUN GOPATH=/tmp/go CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o $GOPATH/bin/pb-testserver ./tools/testserver && \
    GOPATH=/tmp/go CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o $GOPATH/bin/pb ./cmd/pb

######### RUN

FROM alpine
RUN apk update && apk add ca-certificates && apk add tzdata && rm -rf /var/cache/apk/*
COPY --from=builder /go/bin/pb-testserver /usr/local/bin/
COPY --from=builder /go/bin/pb /usr/local/bin/

EXPOSE 5000

ENV PORT 5000
ENV JAEGER_SERVICE_NAME pace-bricks
ENV LOG_FORMAT console

CMD /usr/local/bin/pb-testserver
