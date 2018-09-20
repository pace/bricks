# go-microservice

A tool and kit to help the go developer to stick to pace standards and fasten certain processes.

## Install

    go get -u lab.jamit.de/pace/go-microservice/cmd/pace

## Usage

    pace -h

### Environment variables

#### `PACE_EDITOR`

The path to the editor that should be used for opening a project. Defaults to `$EDITOR`.

#### `PACE_PATH`

The path where new project should be created. Defaults to `$HOME/PACE`.

## Requirements

* A working go installation
* A working git installation
* Installed SSH keys for git

## Testing

Use `make testserver` to test logging and tracing with postgres, redis and external http service.
Use `make docker.all` to create/start all docker containers.

### http/jsonapi

In `http/jsonapi/generator/internal` multiple test APIs can be found. The
generated code in these directories can be updated with `make jsonapi`.

### Jaeger

Access web UI: http://localhost:16686/search

### PostgreSQL

Configure the postgres for go-microservice testserver

    psql -h localhost -U postgres
    > CREATE DATABASE testserver;
    > CREATE USER testserveruser WITH ENCRYPTED PASSWORD 'pace1234!';
    > GRANT ALL PRIVILEGES ON DATABASE testserver TO testserveruser;

### Redis
