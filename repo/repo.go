// Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package repo defines software repository functions like listing, removing
// 	packages from software repository.
package repo

import (
	"flag"
	"fmt"
	logutil "github.com/VeritasOS/plugin-manager/utils/log"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// SoftwareRepoPath is the Software Update Repository path.
const SoftwareRepoPath = "/system/software/repository/"

const (
	// Version of the Software Repository Manager.
	myVersion = "5.2"
)

// cmdOptions contains subcommands and its parameters.
var cmdOptions struct {
	addCmd     *flag.FlagSet
	listCmd    *flag.FlagSet
	removeCmd  *flag.FlagSet
	versionCmd *flag.FlagSet
	versionPtr *bool

	// productVersion indicates the version (i.e., product version) that a
	//  software should be applicable for.
	productVersion string

	// softwareName indicates the name of the software.
	softwareName string

	// softwarePath indicates the path of the software.
	// INFO: This is useful when the software is not part of the repository, and
	// 		one wants get details of software.
	softwarePath string

	// softwareRepo indicates the path to the software repository.
	softwareRepo string

	// softwareType indicates the type of the software.
	softwareType string

	// outputFile indicates the file name to write plugins run results.
	outputFile string

	// outputFormat indicates the output format to write the results.
	//  Supported formats are "json", "yaml".
	outputFormat string
}

// RegisterCommandOptions registers the supported commands.
func RegisterCommandOptions(progname string) {
	log.Printf("Entering repo::RegisterCommandOptions(%s)", progname)
	defer log.Println("Exiting repo::RegisterCommandOptions")

	registerCommandAdd(progname)
	registerCommandList(progname)
	registerCommandRemove(progname)
	registerCommandVersion(progname)
}

func registerCommandVersion(progname string) {
	log.Printf("Entering repo::registerCommandVersion(%s)", progname)
	defer log.Println("Exiting repo::registerCommandVersion")

	cmdOptions.versionCmd = flag.NewFlagSet(progname+" version", flag.ContinueOnError)
	cmdOptions.versionPtr = cmdOptions.versionCmd.Bool(
		"version",
		false,
		"print Software Repository Manager (PM) version.",
	)
}

// ScanCommandOptions scans for the command line options and makes appropriate
// function call.
// Input:
// 	1. map[string]interface{}
//    where, the options could be following:
// 		"progname":  Name of the program along with any cmds (ex: asum pm)
// 		"cmd-index": Index to the cmd (ex: run)
func ScanCommandOptions(options map[string]interface{}) error {
	log.Printf("Entering ScanCommandOptions(%+v)...", options)
	defer log.Println("Exiting ScanCommandOptions")

	progname := filepath.Base(os.Args[0])
	cmdIndex := 1
	if valI, ok := options["progname"]; ok {
		progname = valI.(string)
	}
	if valI, ok := options["cmd-index"]; ok {
		cmdIndex = valI.(int)
	}
	cmd := os.Args[cmdIndex]
	log.Println("progname:", progname, "cmd with arguments:", os.Args[cmdIndex:])

	var err error
	switch cmd {
	case "version":
		logutil.PrintNLog("Software Repository Manager version %s\n", myVersion)

	case "add":
		err = cmdOptions.addCmd.Parse(os.Args[3:])
		if err != nil {
			return logutil.PrintNLogError(cmd, "command arguments parse error:", err.Error())
		}
		err = Add(cmdOptions.softwarePath,
			map[string]string{
				"softwareRepo": cmdOptions.softwareRepo,
			})

	case "list":
		err = cmdOptions.listCmd.Parse(os.Args[3:])
		if err != nil {
			return logutil.PrintNLogError(cmd, "command arguments parse error:", err.Error())
		}

		params := map[string]string{
			"softwareName":   cmdOptions.softwareName,
			"softwareRepo":   cmdOptions.softwareRepo,
			"softwareType":   cmdOptions.softwareType,
			"productVersion": cmdOptions.productVersion,
			"outputFile":     cmdOptions.outputFile,
			"outputFormat":   cmdOptions.outputFormat,
		}
		_, err = List(params)

	case "remove":
		err = cmdOptions.removeCmd.Parse(os.Args[3:])
		if err != nil {
			return logutil.PrintNLogError(cmd, "command arguments parse error:", err.Error())
		}
		err = Remove(cmdOptions.softwareName, cmdOptions.softwareType, cmdOptions.softwareRepo)

	case "help":
		subcmd := ""
		if len(os.Args) == cmdIndex+2 {
			subcmd = os.Args[cmdIndex+1]
		} else if len(os.Args) > cmdIndex+2 {
			fmt.Fprintf(os.Stderr, "usage: %s help command\n\nToo many arguments (%d) given.\n", progname, len(os.Args))
			os.Exit(2)
		}
		usage(progname, subcmd)

	default:
		fmt.Fprintf(os.Stderr, "%s: unknown command \"%s\"\n", progname, os.Args[1])
		fmt.Fprintf(os.Stderr, "Run '%s help [command]' for usage.\n", progname)
		os.Exit(2)
	}

	if err != nil {
		return err
	}
	return nil
}

// Usage of repo command.
func usage(progname, subcmd string) {
	switch subcmd {
	case "", "repo":
		var usageStr = `
Software Repository Manager (PROGNAME ` + subcmd + `) is a tool for managing software repository.

Usage:

	PROGNAME command [arguments]

The commands are:

	add 		add specified software to repository.
	list 		lists contents of software repository.
	remove 		remove specified software from repository.
	version		print Software Repository version.

Use "PROGNAME help [command]" for more information about a command.
		
`
		fmt.Fprintf(os.Stderr, strings.Replace(usageStr, "PROGNAME", progname, -1))
	case "add":
		cmdOptions.addCmd.Usage()
	case "list":
		cmdOptions.listCmd.Usage()
	case "remove":
		cmdOptions.removeCmd.Usage()
	case "version":
		cmdOptions.versionCmd.Usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown help topic `%s`. Run '%s'.", subcmd, progname+" help")
		fmt.Println()
		os.Exit(2)
	}
}
