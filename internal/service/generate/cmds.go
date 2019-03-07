// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/31 by Vincent Landgraf

package generate

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/dave/jennifer/jen"
)

const errorsPkg = "github.com/pace/bricks/maintenance/errors"

// CommandOptions are applied when generating the different
// microservice commands
type CommandOptions struct {
	DaemonName  string
	ControlName string
}

// NewCommandOptions generate command names using given name
func NewCommandOptions(name string) CommandOptions {
	return CommandOptions{
		DaemonName:  name + "d",
		ControlName: name + "ctl",
	}
}

// Commands generates the microservice commands based of
// the given path
func Commands(path string, options CommandOptions) {
	// Create directories
	dirs := []string{
		filepath.Join(path, "cmd", options.DaemonName),
		filepath.Join(path, "cmd", options.ControlName),
	}
	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0770) // nolint: gosec
		if err != nil {
			log.Fatal(fmt.Printf("Failed to create dir %s: %v", dir, err))
		}
	}

	// Create commands files
	for _, dir := range dirs {
		f, err := os.Create(filepath.Join(dir, "main.go"))
		if err != nil {
			log.Fatal(err)
		}

		code := jen.NewFilePathName("", "main")
		cmdName := filepath.Base(dir)

		if cmdName == options.DaemonName {
			generateDaemonMain(code, cmdName)
		} else {
			generateControlMain(code, cmdName)
		}
		_, err = f.WriteString(copyright())
		if err != nil {
			log.Fatal(err)
		}

		_, err = f.WriteString(code.GoString())
		if err != nil {
			log.Fatal(err)
		}
	}
}

func generateDaemonMain(f *jen.File, cmdName string) {
	httpPkg := "github.com/pace/bricks/http"
	logPkg := "github.com/pace/bricks/maintenance/log"
	trancing := "github.com/pace/bricks/maintenance/tracing"

	f.ImportAlias(httpPkg, "pacehttp")
	f.Anon(trancing)
	f.Func().Id("main").Params().BlockFunc(func(g *jen.Group) {
		g.Defer().Qual(errorsPkg, "HandleWithCtx").Call(jen.Qual("context", "Background").Call(), jen.Lit(cmdName))
		g.Id("router").Op(":=").Qual(httpPkg, "Router").Call()
		g.Id("s").Op(":=").Qual(httpPkg, "Server").Call(jen.Id("router"))

		g.Qual(logPkg, "Logger").Call().Dot("Info").Call().Dot("Str").Call(
			jen.Lit("addr"),
			jen.Id("s").Dot("Addr"),
		).Dot("Msg").Call(jen.Lit(fmt.Sprintf("Starting %s ...", cmdName)))

		g.Qual(logPkg, "Fatal").Call(jen.Id("s").Dot("ListenAndServe").Call())
	})
}

func generateControlMain(f *jen.File, cmdName string) {
	f.Func().Id("main").Params().Block(
		jen.Defer().Qual(errorsPkg, "HandleWithCtx").Call(jen.Qual("context", "Background").Call(), jen.Lit(cmdName)),
		jen.Qual("fmt", "Printf").Call(jen.Lit(cmdName)))
}

// copyright generates copyright statement
func copyright() string {
	stmt := ""
	now := time.Now()
	stmt += fmt.Sprintf("// Copyright © %04d by PACE Telematics GmbH. All rights reserved.\n", now.Year())

	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	stmt += fmt.Sprintf("// Created at %04d/%02d/%02d by %s\n\n", now.Year(), now.Month(), now.Day(), u.Name)
	return stmt
}
