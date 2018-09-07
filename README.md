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

### http/jsonapi

In `http/jsonapi/generator/internal` multiple test APIs can be found. The
generated code in these directories can be updated with `make jsonapi`.
