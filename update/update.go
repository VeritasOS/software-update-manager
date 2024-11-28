// Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package update

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	pm "github.com/VeritasOS/plugin-manager" // import "../plugin-manager"
	"github.com/VeritasOS/plugin-manager/config"
	logger "github.com/VeritasOS/plugin-manager/utils/log"
	"github.com/VeritasOS/plugin-manager/utils/output"
	"github.com/VeritasOS/software-update-manager/repo"
	"github.com/VeritasOS/software-update-manager/utils/rpm"
)

// RPMInstallRepoPath is the path where RPM contents are expected to installed/extracted.
const RPMInstallRepoPath = "/system/upgrade/repository/"

// Status of execution used for display as well as logging to specified output file.
const (
	dStatusFail = "Failed"
	dStatusOk   = "Succeeded"
)

// Status is the execution/run status of PM on a specified plugin type.
type Status struct {
	// INFO: The Status contains info of all operations so as to support
	//  all operations with one call to sum with appropriate flags set to
	// 	continue onto the next operation.
	// 	Ex: Install can receive auto-reboot=true, in which after installation
	// 		is completed successfully, `sum` will run reboot operation.
	Install   []pm.Plugin `yaml:",omitempty"`
	Reboot    []pm.Plugin `yaml:",omitempty"`
	Rollback  []pm.Plugin `yaml:",omitempty"`
	Commit    []pm.Plugin `yaml:",omitempty"`
	Status    string
	StdOutErr string
}

// Commit runs the commit-precheck and commit plugins of the update workflow.
func Commit(result *Status, library string) bool {
	logger.Debug.Println("Entering update::Commit")
	defer logger.Debug.Println("Exiting update::Commit")

	// Plugin Types to run for update workflow as part of user commit action.
	pluginTypes := []string{"commit-precheck", "commit"}

	for _, pt := range pluginTypes {
		result.Commit = append(result.Commit, pm.Plugin{})
		resIdx := len(result.Commit) - 1

		err := runPM(&result.Commit[resIdx], pt, library)
		if err != nil {
			return false
		}
	}
	result.Status = dStatusOk
	output.Write(result)
	return true
}

// Install runs the preinstall and install plugins of the update workflow.
func Install(result *Status, library string) bool {
	logger.Debug.Println("Entering update::Install")
	defer logger.Debug.Println("Exiting update::Install")

	// Plugin Types to run for update workflow as part of install script.
	pluginTypes := []string{"preinstall", "install"}
	var err error

	for _, pt := range pluginTypes {
		result.Install = append(result.Install, pm.Plugin{})
		resIdx := len(result.Install) - 1

		err = runPM(&result.Install[resIdx], pt, library)
		if err != nil {
			result.Status = dStatusFail
			break
		}
	}

	if err != nil {
		// INFO: Discard rollback errors, and always return false to
		//  indicate installation failure.
		pt := "rollback"
		result.Install = append(result.Install, pm.Plugin{})
		resIdx := len(result.Install) - 1
		runPM(&result.Install[resIdx], pt, library)
		return false
	}

	result.Status = dStatusOk
	output.Write(result)
	return true
}

// runCmdFromRPM installs the specified software package from the software repo.
func runCmdFromRPM(action, swName, swType string, params map[string]string) error {
	logger.Debug.Printf("Entering update::runCmdFromRPM(%s, %s, %s, %+v)",
		action, swName, swType, params)
	defer logger.Debug.Println("Exiting update::runCmdFromRPM")

	if swName == "" {
		return logger.ConsoleError.PrintNReturnError("Invalid usage. Software name must be specified.")
	}
	if swType == "" {
		return logger.ConsoleError.PrintNReturnError("Invalid usage. Software type must be specified.")
	}
	swRepo := params["softwareRepo"]
	if swRepo == "" {
		swRepo = repo.SoftwareRepoPath
	}

	absSwPath := filepath.Clean(filepath.FromSlash(swRepo +
		string(os.PathSeparator) + swType + string(os.PathSeparator) +
		swName))

	fi, err := os.Stat(absSwPath)
	if err != nil {
		logger.Error.Printf("Unable to stat on %s: %+v, err=%s", absSwPath, fi, err.Error())
		return logger.ConsoleError.PrintNReturnError("Unable to install %s software %s. "+
			"Specified software not found.",
			swType, swName)
	}

	// INFO: If there was an attempt to install this RPM previously,
	// 	then clean up and try installing it again.
	// NOTE: The version validation would have failed if this RPM was
	// 	successfully installed. So no need to worry about removing an RPM
	// 	that was installed successfully.
	params["softwareName"] = swName
	params["softwareType"] = swType
	listInfo, err := repo.List(params)
	if err != nil {
		logger.Error.Printf("Failed to repo.List(), err=%s", err.Error())
		return err
	}
	// Since we had passed softwareName, there is expected to be only one in the list, so get that.
	if 1 != len(listInfo) {
		logger.Error.Printf("The repo list is expected to have details of %s software, but got %+v.", swName, listInfo)
		return logger.ConsoleError.PrintNReturnError("Failed to get details of %s software.", swName)
	}
	rpmInfo := listInfo[0]

	// INFO: Only for `install` action, we need to first install/extract the
	// 		SUM RPM to get the install script. For all other actions, the
	// 		install script is expected to run first, and hence they're
	// 		expected to be present at the scripts location.
	if "install" == action {
		if rpm.IsInstalled(rpmInfo.GetRPMName()) {
			rpm.Uninstall(rpmInfo.GetRPMName())
		}

		err = rpm.Install(absSwPath)
		if err != nil {
			return logger.ConsoleError.PrintNReturnError("Failed to install software.")
		}
	}

	script := RPMInstallRepoPath + swType + string(os.PathSeparator) +
		fmt.Sprintf("%s-%s-%s", rpmInfo.GetRPMName(),
			rpmInfo.GetRPMVersion(), rpmInfo.GetRPMRelease()) +
		string(os.PathSeparator) + action
	logger.Info.Println("Script to be invoked:", script)

	const cmdStr = "/bin/sh"
	cmdParams := []string{script, "-output-file", output.GetFile(),
		"-output-format", output.GetFormat()}
	cmd := exec.Command(os.ExpandEnv(cmdStr), cmdParams...)
	stdOutErr, err := cmd.CombinedOutput()
	logger.Debug.Println("Stdout & Stderr:", string(stdOutErr))
	if err != nil {
		logger.Error.Printf("Failed to run %s script of %s RPM, err=%s", script, absSwPath, err.Error())
		if "install" == action {
			rpm.Uninstall(rpmInfo.GetRPMName())
		}
		return logger.ConsoleError.PrintNReturnError("Failed to %s software.", action)
	}

	logger.Info.Printf("Successfully completed %s of %s software", action, absSwPath)
	logger.ConsoleInfo.Printf("Successfully completed %s of %s software %s from repository.\n",
		action, swType, swName)
	return nil
}

// Reboot runs the prereboot plugins and reboot the system as part of the update workflow.
func Reboot(result *Status, library string) bool {
	logger.Debug.Println("Entering update::Reboot")
	defer logger.Debug.Println("Exiting update::Reboot")

	// Plugin Types to run for update workflow as part of prereboot script.
	pluginTypes := []string{"prereboot"}

	for _, pt := range pluginTypes {
		result.Reboot = append(result.Reboot, pm.Plugin{})
		resIdx := len(result.Reboot) - 1

		err := runPM(&result.Reboot[resIdx], pt, library)
		if err != nil {
			result.Status = dStatusFail
			result.StdOutErr = err.Error()
			// INFO: Discard rollback errors, and always return false to
			//  indicate reboot failure.
			pt := "rollback"
			result.Reboot = append(result.Reboot, pm.Plugin{})
			resIdx := len(result.Reboot) - 1

			runPM(&result.Reboot[resIdx], pt, library)
			return false
		}
	}
	result.Status = dStatusOk
	output.Write(result)

	// Reboot the system after prereboot plugins are run successfully.
	cmdStr := "systemctl"
	cmdParams := []string{"reboot"}

	cmd := exec.Command(os.ExpandEnv(cmdStr), cmdParams...)
	stdOutErr, err := cmd.CombinedOutput()
	logger.Debug.Println("Stdout & Stderr:", string(stdOutErr))
	if err != nil {
		logger.Error.Printf("Failed to reboot the system, err=%s", err.Error())
	}

	return true
}

func runPM(result *pm.Plugin, pluginType, library string) error {
	logger.Debug.Println("Entering update::runPM")
	defer logger.Debug.Println("Exiting update::runPM")

	logger.ConsoleInfo.Printf("Running %s plugins...", pluginType)
	config.SetPluginsLibrary(library)

	err := pm.RunFromLibrary(result, pluginType, pm.RunOptions{Library: library})
	if err != nil {
		logger.Error.Printf("Failed to run %s plugins, err=%s", pluginType, err.Error())
		return err
	}
	fmt.Println()
	return nil
}

// Rollback runs the required rollback plugins of the update workflow in the new version/partition.
func Rollback(result *Status, library string) bool {
	logger.Debug.Println("Entering update::Rollback")
	defer logger.Debug.Println("Exiting update::Rollback")

	// Plugin Types to run for update workflow as part of rollback script.
	pluginTypes := []string{"rollback-precheck", "prerollback"}

	for _, pt := range pluginTypes {
		result.Rollback = append(result.Rollback, pm.Plugin{})
		resIdx := len(result.Rollback) - 1

		err := runPM(&result.Rollback[resIdx], pt, library)
		if err != nil {
			return false
		}
	}
	result.Status = dStatusOk
	output.Write(result)
	return true
}

// cmdOptions contains subcommands and parameters of the pm command.
var cmdOptions struct {
	commitCmd   *flag.FlagSet
	installCmd  *flag.FlagSet
	rebootCmd   *flag.FlagSet
	rollbackCmd *flag.FlagSet

	// softwareName indicates the name of the software.
	softwareName string

	// softwareRepo indicates the path to the software repository.
	softwareRepo string

	// softwareType indicates the type of the software.
	softwareType string

	// logDir indicates the location for writing log file.
	logDir string

	// logFile indicates the log file name to write to in the logDir location.
	logFile string
}

func registerCmdOptions(f *flag.FlagSet) {
	f.StringVar(
		&cmdOptions.softwareName,
		"filename",
		"",
		"File name of the software.",
	)
	f.StringVar(
		&cmdOptions.softwareRepo,
		"repo",
		"",
		"Path of the software repository.",
	)
	f.StringVar(
		&cmdOptions.softwareType,
		"type",
		"",
		"Type of the software.",
	)
	output.RegisterCommandOptions(f, map[string]string{"output-format": "yaml"})
}

// registerCommandCommit registers install command and its options
func registerCommandCommit(progname string) {
	logger.Debug.Printf("Entering update::registerCommandCommit(%s)", progname)
	defer logger.Debug.Println("Exiting update::registerCommandCommit")

	cmdOptions.commitCmd = flag.NewFlagSet(progname+" commit", flag.PanicOnError)
	registerCmdOptions(cmdOptions.commitCmd)
}

// registerCommandInstall registers install command and its options
func registerCommandInstall(progname string) {
	logger.Debug.Printf("Entering update::registerCommandInstall(%s)", progname)
	defer logger.Debug.Println("Exiting update::registerCommandInstall")

	cmdOptions.installCmd = flag.NewFlagSet(progname+" install", flag.PanicOnError)
	registerCmdOptions(cmdOptions.installCmd)
}

// registerCommandReboot registers reboot command and its options
func registerCommandReboot(progname string) {
	logger.Info.Printf("Entering update::registerCommandReboot(%s)", progname)
	defer logger.Info.Println("Exiting update::registerCommandReboot")

	cmdOptions.rebootCmd = flag.NewFlagSet(progname+" reboot", flag.PanicOnError)
	registerCmdOptions(cmdOptions.rebootCmd)
}

// registerCommandRollback registers rollback command and its options
func registerCommandRollback(progname string) {
	logger.Debug.Printf("Entering update::registerCommandRollback(%s)", progname)
	defer logger.Debug.Println("Exiting update::registerCommandRollback")

	cmdOptions.rollbackCmd = flag.NewFlagSet(progname+" rollback", flag.PanicOnError)
	registerCmdOptions(cmdOptions.rollbackCmd)
}

// RegisterCommandOptions registers the command options that are supported
func RegisterCommandOptions(progname string) {
	logger.Debug.Println("Entering update::RegisterCommandOptions")
	defer logger.Debug.Println("Exiting update::RegisterCommandOptions")

	registerCommandCommit(progname)
	registerCommandInstall(progname)
	registerCommandReboot(progname)
	registerCommandRollback(progname)
}

// ScanCommandOptions scans for the command line options and makes appropriate
// function call.
// Input:
//  1. map[string]interface{}
//     where, the options could be following:
//     "progname":  Name of the program along with any cmds (ex: sum pm)
//     "cmd-index": Index to the cmd (ex: run)
func ScanCommandOptions(options map[string]interface{}) error {
	logger.Debug.Printf("Entering ScanCommandOptions(%+v)...", options)
	defer logger.Debug.Println("Exiting ScanCommandOptions")

	progname := filepath.Base(os.Args[0])
	cmdIndex := 1
	if valI, ok := options["progname"]; ok {
		progname = valI.(string)
	}
	if valI, ok := options["cmd-index"]; ok {
		cmdIndex = valI.(int)
	}
	cmd := os.Args[cmdIndex]
	logger.Debug.Println("Progname: ", progname, " cmd with arguments: ", os.Args[cmdIndex:])

	var err error
	switch cmd {
	case "commit":
		err = cmdOptions.commitCmd.Parse(os.Args[cmdIndex+1:])

	case "install":
		err = cmdOptions.installCmd.Parse(os.Args[cmdIndex+1:])

	case "reboot":
		err = cmdOptions.rebootCmd.Parse(os.Args[cmdIndex+1:])

	case "rollback":
		err = cmdOptions.rollbackCmd.Parse(os.Args[cmdIndex+1:])

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
		fmt.Fprintf(os.Stderr, "%s: unknown command \"%s\"\n", progname, cmd)
		fmt.Fprintf(os.Stderr, "Run '%s help [command]' for usage.\n", progname)
		os.Exit(2)
	}

	if err != nil {
		return logger.ConsoleError.PrintNReturnError("Command arguments parse error, cmd=%s, err=%s", cmd, err.Error())
	}

	if cmdOptions.softwareName != "" {
		params := map[string]string{}
		params["softwareRepo"] = cmdOptions.softwareRepo
		err = runCmdFromRPM(cmd, cmdOptions.softwareName, cmdOptions.softwareType, params)
	} else {
		library := options["library"].(string)

		var status Status
		var ret bool
		switch cmd {
		case "commit":
			ret = Commit(&status, library)
		case "install":
			ret = Install(&status, library)
		case "reboot":
			ret = Reboot(&status, library)
		case "rollback":
			ret = Rollback(&status, library)
		}
		if !ret {
			err = logger.ConsoleError.PrintNReturnError("Failed to %s the update.", cmd)
			status.Status = dStatusFail
			status.StdOutErr = err.Error()
			output.Write(status)
		}
	}
	return err
}

// Usage of command.
func usage(progname, subcmd string) {
	switch subcmd {
	case "commit":
		cmdOptions.commitCmd.Usage()
	case "install":
		cmdOptions.installCmd.Usage()
	case "reboot":
		cmdOptions.rebootCmd.Usage()
	case "rollback":
		cmdOptions.rollbackCmd.Usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown help topic `%s`. Run '%s'.", subcmd, progname+" help")
		fmt.Println()
		os.Exit(2)
	}
}
