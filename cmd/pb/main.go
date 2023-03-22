// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/10 by Vincent Landgraf

package main

import (
	"github.com/spf13/cobra"

	"github.com/pace/bricks/internal/service"
	"github.com/pace/bricks/internal/service/generate"
	"github.com/pace/bricks/maintenance/log"
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
	var restSource string
	rootCmdNew := &cobra.Command{
		Use:  "new NAME",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.New(args[0], service.NewOptions{
				RestSource: restSource,
			})
		},
	}
	rootCmdNew.Flags().StringVar(&restSource, "source", "", "OpenAPIv3 source (URI / path) to use for generation")
	rootCmd.AddCommand(rootCmdNew)

	rootCmdClone := &cobra.Command{
		Use:  "clone NAME",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Clone(args[0])
		},
	}
	rootCmd.AddCommand(rootCmdClone)

	rootCmdPath := &cobra.Command{
		Use:  "path NAME",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Path(args[0])
		},
	}
	rootCmd.AddCommand(rootCmdPath)

	rootCmdEdit := &cobra.Command{
		Use:  "edit NAME",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Edit(args[0])
		},
	}
	rootCmd.AddCommand(rootCmdEdit)

	var runCmd string
	rootCmdRun := &cobra.Command{
		Use:  "run NAME",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Run(args[0], service.RunOptions{
				CmdName: runCmd,
				Args:    args,
			})
		},
	}
	rootCmdRun.Flags().StringVar(&runCmd, "cmd", "", "name of the command to run")
	rootCmd.AddCommand(rootCmdRun)

	var testGoConvey bool
	rootCmdTest := &cobra.Command{
		Use:  "test NAME",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Test(args[0], service.TestOptions{GoConvey: testGoConvey})
		},
	}
	rootCmdTest.Flags().BoolVar(&testGoConvey, "goconvey", false, "use goconvey for testing")
	rootCmd.AddCommand(rootCmdTest)

	rootCmdLint := &cobra.Command{
		Use:  "lint NAME",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Lint(args[0])
		},
	}
	rootCmd.AddCommand(rootCmdLint)

	rootCmdGenerate := &cobra.Command{
		Use:  "generate [command]",
		Args: cobra.MaximumNArgs(1),
	}
	rootCmd.AddCommand(rootCmdGenerate)
	addServiceGenerateCommands(rootCmdGenerate)
}

// pace service generate ...
func addServiceGenerateCommands(rootCmdGenerate *cobra.Command) {
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
	rootCmdGenerate.AddCommand(cmdRest)

	var commandsPath string
	cmdCommands := &cobra.Command{
		Use:  "commands NAME",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			generate.Commands(commandsPath,
				generate.NewCommandOptions(args[0]))
		},
	}
	cmdCommands.Flags().StringVar(&commandsPath, "path", "", "path directory in which to create the commands")
	rootCmdGenerate.AddCommand(cmdCommands)

	var dockerfilePath string
	cmdDockerfile := &cobra.Command{
		Use:  "dockerfile NAME",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			generate.Dockerfile(dockerfilePath, generate.DockerfileOptions{
				Name:     args[0],
				Commands: generate.NewCommandOptions(args[0]),
			})
		},
	}
	cmdDockerfile.Flags().StringVar(&dockerfilePath, "path", "./Dockerfile", "path to Dockerfile location")
	rootCmdGenerate.AddCommand(cmdDockerfile)

	var makefilePath string
	cmdMakefile := &cobra.Command{
		Use:  "makefile NAME",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			generate.Makefile(makefilePath, generate.MakefileOptions{
				Name: args[0],
			})
		},
	}
	cmdMakefile.Flags().StringVar(&makefilePath, "path", "./Makefile", "path to Makefile location")
	rootCmdGenerate.AddCommand(cmdMakefile)

	var errorsDefinitionsPkgName, errorsDefinitionsPath, errorsDefinitionsSource string

	cmdErrorDefinitions := &cobra.Command{
		Use:   "error-definitions",
		Short: "generate BricksErrors based on an array of JSON error objects",
		Run: func(cmd *cobra.Command, args []string) {
			generate.ErrorDefinitionFile(generate.ErrorDefinitionFileOptions{
				PkgName: errorsDefinitionsPkgName,
				Path:    errorsDefinitionsPath,
				Source:  errorsDefinitionsSource,
			})
		},
	}
	cmdErrorDefinitions.Flags().StringVar(&errorsDefinitionsPkgName, "pkg", "", "name for the generated go package")
	cmdErrorDefinitions.Flags().StringVar(&errorsDefinitionsPath, "path", "", "path for generated file")
	cmdErrorDefinitions.Flags().StringVar(&errorsDefinitionsSource, "source", "", "JSONAPI conform error definitions source to use for generation")
	rootCmdGenerate.AddCommand(cmdErrorDefinitions)
	addErrorDefinitionsCommands(cmdErrorDefinitions)
}

func addErrorDefinitionsCommands(rootCmdErrorDefinitions *cobra.Command) {
	var inputPath, outputPath string

	cmdMarkdown := &cobra.Command{
		Use:   "markdown",
		Short: "generate error definitions markdown from yaml file",
		Run: func(cmd *cobra.Command, args []string) {
			generate.ErrorDefinitionsMarkdown(inputPath, outputPath)
		},
	}
	cmdMarkdown.Flags().StringVarP(&inputPath, "input", "i", "", "path for the yaml file")
	cmdMarkdown.Flags().StringVarP(&outputPath, "output", "o", "", "path for the generated markdown file")
	rootCmdErrorDefinitions.AddCommand(cmdMarkdown)
}
