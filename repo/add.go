// Copyright (c) 2022 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package repo defines software repository functions like listing, removing
//
//	packages from software repository.
package repo

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/VeritasOS/plugin-manager/config"
	logger "github.com/VeritasOS/plugin-manager/utils/log"
	osutils "github.com/VeritasOS/plugin-manager/utils/os"
)

// Add the software file present in the staging area to the software repo after
//
//	validation.
func Add(rpmPath string, params map[string]string) error {
	logger.Debug.Printf("Entering repo::Add(%v, %v)", rpmPath, params)
	defer logger.Debug.Println("Exiting repo::Add")

	productVersion := params["productVersion"]
	swRepo := params["softwareRepo"]

	fi, err := os.Stat(rpmPath)
	if err != nil {
		return logger.ConsoleError.PrintNReturnError(
			"Unable to stat on %s software. Error: %s\n",
			rpmPath, err.Error())
	} else if fi.IsDir() {
		return logger.ConsoleError.PrintNReturnError(
			"%s is not a valid software.\n",
			rpmPath)
	}

	info, err := ListRPMFilesInfo([]string{rpmPath}, productVersion)
	if err != nil {
		return err
	}

	// As listing was done for one RPM, the list is expected to have just one RPM.
	rpmType := info[0].GetRPMType()
	if "" == rpmType {
		return logger.ConsoleError.PrintNReturnError("Failed to determine the software type of the %s file.",
			rpmPath)
	}
	rpmType = strings.ToLower(rpmType)

	repoTypeLocation := filepath.FromSlash(swRepo + "/" + rpmType)
	if err := osutils.OsMkdirAll(repoTypeLocation, 0755); nil != err {
		return logger.ConsoleError.PrintNReturnError("Failed to create the plugins logs directory: %s. "+
			"Error: %s", config.GetPluginsLogDir(), err.Error())
	}

	const cmdStr = "/usr/bin/mv"
	cmdParams := []string{"-f", rpmPath, repoTypeLocation}
	cmd := exec.Command(os.ExpandEnv(cmdStr), cmdParams...)
	stdOutErr, err := cmd.CombinedOutput()
	logger.Debug.Println("Stdout & Stderr:", string(stdOutErr))
	if err != nil {
		logger.Error.Printf("Failed to move %s RPM to software update repository %s, err=%s",
			rpmPath, repoTypeLocation, err.Error())
		return logger.ConsoleError.PrintNReturnError("Failed to add %s software to "+
			"software repository.", rpmPath)
	}

	return nil
}

//	 registerCommandAdd registers the add command that enables one to
//		add the RPM to the software update repository.
func registerCommandAdd(progname string) {
	logger.Debug.Printf("Entering repo::registerCommandAdd(%s)", progname)
	defer logger.Debug.Println("Exiting repo::registerCommandAdd")

	cmdOptions.addCmd = flag.NewFlagSet(progname+" add", flag.PanicOnError)

	cmdOptions.addCmd.StringVar(
		&cmdOptions.productVersion,
		"product-version",
		"",
		"Version that a software should be compatibile with."+
			" (I.e., product-version)",
	)
	cmdOptions.addCmd.StringVar(
		&cmdOptions.softwarePath,
		"filepath",
		"",
		"Path of the software.",
	)
	cmdOptions.addCmd.StringVar(
		&cmdOptions.softwareRepo,
		"repo",
		SoftwareRepoPath,
		"Path of the software repository.",
	)
	cmdOptions.addCmd.StringVar(
		&cmdOptions.outputFile,
		"output-file",
		"",
		"Name of the file to write the results.",
	)
	cmdOptions.addCmd.StringVar(
		&cmdOptions.outputFormat,
		"output-format",
		"yaml",
		"The format of output to display the results. "+
			"Supported output formats are 'json', 'yaml'.",
	)
}
