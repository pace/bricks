# PACE Bricks

![](artwork/PACE-Bricks_Header_LightBackground.png)

Opinionated microservice kit to help developers to build microservices with go.

## Install

    go get github.com/pace/bricks/cmd/pb

## Usage

    pb -h

### Environment variables

#### `PACE_BRICKS_EDITOR`

The path to the editor that should be used for opening a project. Defaults to `$EDITOR`.

#### `PACE_BRICKS_PATH`

The path where new project should be created. Defaults to `$HOME/PACE`.

## Requirements

* A working go installation
* A working git installation
* Installed SSH keys for git

## Testing

Use `make testserver` to test logging and tracing with postgres, redis and external http service.
Use `make docker.all` to create/start all docker containers.

## Configuration

All of the microservices follow the [TWELVE-FACTOR APP](https://12factor.net/) standard of environment based configuration.

### http/jsonapi

In `http/jsonapi/generator/internal` multiple test APIs can be found. The
generated code in these directories can be updated with `make jsonapi`.

### Jaeger

Access web UI: http://localhost:16686/search
