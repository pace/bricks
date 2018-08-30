# PACE tool

A tool to help the go developer to stick to pace standards and fasten certain processes.

## Install

    go get -u lab.jamit.de/pace/tool

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
