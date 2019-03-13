// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/10 by Vincent Landgraf

package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pace/bricks/internal/service"
	"github.com/pace/bricks/internal/service/generate"
	"github.com/pace/bricks/maintenance/log"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:  "pb [command]",
		Args: cobra.MaximumNArgs(1),
	}
	addRootCommands(rootCmd)
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

// pace ...
func addRootCommands(rootCmd *cobra.Command) {
	cmdService := &cobra.Command{
		Use:   "service [command]",
		Short: "All service related commands",
		Args:  cobra.MaximumNArgs(1),
	}
	rootCmd.AddCommand(cmdService)
	addServiceCommands(cmdService)

	var runCmd string
	cmdRun := &cobra.Command{
		Use:   "run",
		Short: "Starts the correct server from the current directory",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Run(mustIdentifyService(), service.RunOptions{
				CmdName: runCmd,
				Args:    args,
			})
		},
	}
	cmdRun.Flags().StringVar(&runCmd, "cmd", "", "name of the command to run")
	rootCmd.AddCommand(cmdRun)

	cmdLint := &cobra.Command{
		Use:   "lint",
		Short: "Lints the correct code of the current directory",
		Run: func(cmd *cobra.Command, args []string) {
			service.Lint(mustIdentifyService())
		},
	}
	rootCmd.AddCommand(cmdLint)

	var testGoConvey bool
	cmdTest := &cobra.Command{
		Use:   "test",
		Short: "Tests the correct code of the currect directory",
		Run: func(cmd *cobra.Command, args []string) {
			service.Test(mustIdentifyService(), service.TestOptions{GoConvey: testGoConvey})
		},
	}
	cmdTest.Flags().BoolVar(&testGoConvey, "goconvey", false, "use goconvey for testing")
	rootCmd.AddCommand(cmdTest)
}

// pace service ...
func addServiceCommands(cmdService *cobra.Command) {
	var restSource string
	cmdServiceNew := &cobra.Command{
		Use:  "new [name]",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.New(args[0], service.NewOptions{
				RestSource: restSource,
			})
		},
	}
	cmdServiceNew.Flags().StringVar(&restSource, "source", "", "OpenAPIv3 source (URI / path) to use for generation")
	cmdService.AddCommand(cmdServiceNew)

	cmdServiceClone := &cobra.Command{
		Use:  "clone [name]",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Clone(args[0])
		},
	}
	cmdService.AddCommand(cmdServiceClone)

	cmdServicePath := &cobra.Command{
		Use:  "path [name]",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Path(args[0])
		},
	}
	cmdService.AddCommand(cmdServicePath)

	cmdServiceEdit := &cobra.Command{
		Use:  "edit [name]",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Edit(args[0])
		},
	}
	cmdService.AddCommand(cmdServiceEdit)

	var runCmd string
	cmdServiceRun := &cobra.Command{
		Use:  "run [name]",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Run(args[0], service.RunOptions{
				CmdName: runCmd,
				Args:    args,
			})
		},
	}
	cmdServiceRun.Flags().StringVar(&runCmd, "cmd", "", "name of the command to run")
	cmdService.AddCommand(cmdServiceRun)

	var testGoConvey bool
	cmdServiceTest := &cobra.Command{
		Use:  "test [name]",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Test(args[0], service.TestOptions{GoConvey: testGoConvey})
		},
	}
	cmdServiceTest.Flags().BoolVar(&testGoConvey, "goconvey", false, "use goconvey for testing")
	cmdService.AddCommand(cmdServiceTest)

	cmdServiceLint := &cobra.Command{
		Use:  "lint [name]",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Lint(args[0])
		},
	}
	cmdService.AddCommand(cmdServiceLint)

	cmdServiceGenerate := &cobra.Command{
		Use:  "generate [command]",
		Args: cobra.MaximumNArgs(1),
	}
	cmdService.AddCommand(cmdServiceGenerate)
	addServiceGenerateCommands(cmdServiceGenerate)
}

// pace service generate ...
func addServiceGenerateCommands(cmdServiceGenerate *cobra.Command) {
	var pkgName, path, source string
	cmdRest := &cobra.Command{
		Use:  "rest",
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			generate.Rest(generate.RestOptions{
				PkgName: pkgName,
				Path:    path,
				Source:  source,
			})
		},
	}
	cmdRest.Flags().StringVar(&pkgName, "pkg", "", "name for the generated go package")
	cmdRest.Flags().StringVar(&path, "path", "", "path for generated file")
	cmdRest.Flags().StringVar(&source, "source", "", "OpenAPIv3 source to use for generation")
	cmdServiceGenerate.AddCommand(cmdRest)

	var commandsPath string
	cmdCommands := &cobra.Command{
		Use:  "commands [name]",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			generate.Commands(commandsPath,
				generate.NewCommandOptions(args[0]))
		},
	}
	cmdCommands.Flags().StringVar(&commandsPath, "path", "", "path directory in which to create the commands")
	cmdServiceGenerate.AddCommand(cmdCommands)

	var dockerfilePath string
	cmdDockerfile := &cobra.Command{
		Use:  "dockerfile [name]",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			generate.Dockerfile(dockerfilePath, generate.DockerfileOptions{
				Name:     args[0],
				Commands: generate.NewCommandOptions(args[0]),
			})
		},
	}
	cmdDockerfile.Flags().StringVar(&dockerfilePath, "path", "./Dockerfile", "path to Dockerfile location")
	cmdServiceGenerate.AddCommand(cmdDockerfile)

	var makefilePath string
	cmdMakefile := &cobra.Command{
		Use:  "makefile [name]",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			generate.Makefile(makefilePath, generate.MakefileOptions{
				Name: args[0],
			})
		},
	}
	cmdMakefile.Flags().StringVar(&makefilePath, "path", "./Makefile", "path to Makefile location")
	cmdServiceGenerate.AddCommand(cmdMakefile)
}

// mustIdentifyService returns the name of the current service or quits the program
func mustIdentifyService() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// find relation of current path and the gopath (if we are not in the gopath -> err)
	relDir, err := filepath.Rel(service.GoPath(), dir)
	if err != nil {
		log.Fatal(err)
	}

	// check if the current path is a service
	servicePrefix := filepath.Join("src", service.PaceBase, service.ServiceBase)
	if !strings.HasPrefix(relDir, servicePrefix) {
		log.Fatalf("%s is not a service directory", dir)
	}

	// err if on service top level dir
	if servicePrefix == relDir {
		log.Fatalf("%s is not a service directory", dir)
	}

	// trim the gopath service prefix from the project
	base := strings.TrimPrefix(relDir, servicePrefix+"/")
	if err != nil {
		log.Fatal(err)
	}

	// top level of project return current name
	if filepath.Base(base) == base {
		return base
	}

	// when in sub directory return the base (service name)
	return filepath.Dir(base)
}
