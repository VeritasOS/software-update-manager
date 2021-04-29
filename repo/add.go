// Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package repo defines software repository functions like listing, removing
// 	packages from software repository.
package repo

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"plugin-manager/config"
	logutil "plugin-manager/utils/log"
	osutils "plugin-manager/utils/os"
	"strings"
)

// Add the software file present in the staging area to the software repo after
// 	validation.
func Add(rpmPath string, params map[string]string) error {
	log.Printf("Entering repo::Add(%v, %v)", rpmPath, params)
	defer log.Println("Exiting repo::Add")

	productVersion := params["productVersion"]
	swRepo := params["softwareRepo"]

	fi, err := os.Stat(rpmPath)
	if err != nil {
		return logutil.PrintNLogError(
			"Unable to stat on %s software. Error: %s\n",
			rpmPath, err.Error())
	} else if fi.IsDir() {
		return logutil.PrintNLogError(
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
		return logutil.PrintNLogError("Failed to determine the software type of the %s file.",
			rpmPath)
	}
	rpmType = strings.ToLower(rpmType)

	repoTypeLocation := filepath.FromSlash(swRepo + "/" + rpmType)
	if err := osutils.OsMkdirAll(repoTypeLocation, 0755); nil != err {
		return logutil.PrintNLogError("Failed to create the plugins logs directory: %s. "+
			"Error: %s", config.GetPluginsLogDir(), err.Error())
	}

	const cmdStr = "/usr/bin/mv"
	cmdParams := []string{"-f", rpmPath, repoTypeLocation}
	cmd := exec.Command(os.ExpandEnv(cmdStr), cmdParams...)
	stdOutErr, err := cmd.CombinedOutput()
	log.Println("Stdout & Stderr:", string(stdOutErr))
	if err != nil {
		log.Printf("Failed to move %s RPM to software update repository %s. Error: %s\n",
			rpmPath, repoTypeLocation, err.Error())
		return logutil.PrintNLogError("Failed to add %s software to "+
			"software repository.", rpmPath)
	}

	return nil
}

//  registerCommandAdd registers the add command that enables one to
// 	add the RPM to the software update repository.
func registerCommandAdd(progname string) {
	log.Printf("Entering repo::registerCommandAdd(%s)", progname)
	defer log.Println("Exiting repo::registerCommandAdd")

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
