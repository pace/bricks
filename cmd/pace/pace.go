package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"lab.jamit.de/pace/web/tool/internal/service"
)

func main() {
	rootCmd := &cobra.Command{
		Use:  "pace [command]",
		Args: cobra.MaximumNArgs(1),
	}
	addRootCommands(rootCmd)
	rootCmd.Execute()
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
	cmdServiceNew := &cobra.Command{
		Use:  "new",
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.New(args[0])
		},
	}
	cmdService.AddCommand(cmdServiceNew)

	cmdServiceClone := &cobra.Command{
		Use:  "clone",
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Clone(args[0])
		},
	}
	cmdService.AddCommand(cmdServiceClone)

	cmdServicePath := &cobra.Command{
		Use:  "path",
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Path(args[0])
		},
	}
	cmdService.AddCommand(cmdServicePath)

	cmdServiceEdit := &cobra.Command{
		Use:  "edit",
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Edit(args[0])
		},
	}
	cmdService.AddCommand(cmdServiceEdit)

	var runCmd string
	cmdServiceRun := &cobra.Command{
		Use:  "run",
		Args: cobra.MinimumNArgs(1),
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
		Use:  "test",
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Test(args[0], service.TestOptions{GoConvey: testGoConvey})
		},
	}
	cmdServiceTest.Flags().BoolVar(&testGoConvey, "goconvey", false, "use goconvey for testing")
	cmdService.AddCommand(cmdServiceTest)

	cmdServiceLint := &cobra.Command{
		Use:  "lint",
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service.Lint(args[0])
		},
	}
	cmdService.AddCommand(cmdServiceLint)
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
	if !filepath.HasPrefix(relDir, servicePrefix) {
		log.Fatalf("%s is not a service directory", dir)
	}

	// err if on service top level dir
	if servicePrefix == relDir {
		log.Fatalf("%s is not a service directory", dir)
	}

	// trim the gopath service prefix from the project
	base := strings.TrimPrefix(relDir, servicePrefix+"/")
	log.Print(base)
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
