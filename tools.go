// +build tools

package bricks

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/mattn/goveralls"
	_ "golang.org/x/tools/cmd/cover"
)
