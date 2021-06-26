// Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package main

import (
	"flag"
	"fmt"
	pm "github.com/VeritasOS/plugin-manager" // import "../../plugin-manager"
	"github.com/VeritasOS/plugin-manager/config"
	logutil "github.com/VeritasOS/plugin-manager/utils/log"
	"github.com/VeritasOS/software-update-manager/repo"
	"github.com/VeritasOS/software-update-manager/update"
	"github.com/VeritasOS/software-update-manager/validate"
	"os"
	"path/filepath"
	"strings"
)

var (
	buildDate string
	// progname is name of my binary/program/executable.
	progname = filepath.Base(os.Args[0])
	// version of my program.
	version = "5.9"
)

func mainRegisterCmdOptions() {
	mainCmdOptions.versionCmd = flag.NewFlagSet(progname+" version", flag.ContinueOnError)
	mainCmdOptions.versionPtr = mainCmdOptions.versionCmd.Bool("version", false, "print Plugin Manager version.")
}

var mainCmdOptions struct {
	versionCmd *flag.FlagSet
	versionPtr *bool
}

func init() {
	config.SetLogDir("/var/log/sum/")
}

func main() {
	absprogpath, err := filepath.Abs(os.Args[0])
	if err != nil {
		logutil.PrintNLogError("Failed to get the %s path.", progname)
	}

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Subcommand as operation is required.\n")
		os.Exit(1)
	}

	cmd := os.Args[1]
	config.SetLogFile(cmd)
	logutil.SetLogging(config.GetLogDir() + config.GetLogFile())

	mainRegisterCmdOptions()
	pm.RegisterCommandOptions(progname + " pm")
	repo.RegisterCommandOptions(progname + " repo")
	update.RegisterCommandOptions(progname)
	switch cmd {
	case "version":
		logutil.PrintNLog("%s version %s %s\n", progname, version, buildDate)

	case "commit", "install", "reboot", "rollback":
		library := filepath.Clean(
			filepath.Dir(absprogpath) + string(os.PathSeparator) + "library")
		err := update.ScanCommandOptions(map[string]interface{}{"library": library})
		if nil != err {
			os.Exit(1)
		}

	case "pm":
		options := map[string]interface{}{
			"progname":  progname + " " + cmd,
			"cmd-index": 2,
		}
		err := pm.ScanCommandOptions(options)
		if err != nil {
			os.Exit(1)
		}

	case "repo":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Subcommand %s requires arguments.\n",
				cmd)
			os.Exit(1)
		}
		options := map[string]interface{}{
			"progname":  progname + " " + cmd,
			"cmd-index": 2,
		}
		err = repo.ScanCommandOptions(options)
		if err != nil {
			os.Exit(1)
		}

	case "validate":
		validate.Exec(os.Args[1:])

	case "help":
		subcmd := ""
		if len(os.Args) == 3 {
			subcmd = os.Args[2]
		} else if len(os.Args) > 3 {
			fmt.Fprintf(os.Stderr, "usage: %s help command\n\nToo many arguments (%d) given.\n", progname, len(os.Args))
			os.Exit(2)
		}
		usage(progname, subcmd)

	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s.\n", os.Args[1])
		os.Exit(1)
	}
}

// Usage of Plugin Manager (pm) command.
func usage(progname, subcmd string) {
	switch subcmd {
	case "":
		const usageStr = `
SUM (PROGNAME) is a tool for Software Updates Management (SUM).

Usage:

	PROGNAME command [arguments]

The commands are:

	commit		commits the installed software update.
	install		installs software update.
	pm   		perform Plugin Manager (PM) operations.
	reboot		reboots/restarts the node running reboots specific action for installing software update.
	repo 		perform Software Repository management operations.
	rollback	rolls back the installed software update.
	version		print Software Updates Management (SUM) version.

Use "PROGNAME help [command]" for more information about a command.
		
`
		fmt.Fprintf(os.Stderr, strings.Replace(usageStr, "PROGNAME", progname, -1))
	case "version":
		mainCmdOptions.versionCmd.Usage()

	case "commit", "install", "reboot", "rollback":
		update.ScanCommandOptions(nil)

	case "pm":
		pm.ScanCommandOptions(nil)

	case "repo":
		repo.ScanCommandOptions(nil)

	default:
		fmt.Fprintf(os.Stderr, "Unknown help topic `%s`. Run '%s'.", subcmd, progname+" help")
		fmt.Println()
		os.Exit(2)
	}
}
