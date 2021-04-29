// Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package repo defines software repository functions like listing, removing
// 	packages from software repository.
package repo

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	logutil "plugin-manager/utils/log"
	osutils "plugin-manager/utils/os"
	"strings"
)

// registerCommandRemove registers the remove command that enables one to
// 	remove the RPM of the specified type from the software update repository.
func registerCommandRemove(progname string) {
	log.Printf("Entering repo::registerCommandRemove(%s)", progname)
	defer log.Println("Exiting repo::registerCommandRemove")

	cmdOptions.removeCmd = flag.NewFlagSet(progname+" remove", flag.PanicOnError)
	cmdOptions.removeCmd.StringVar(
		&cmdOptions.softwareRepo,
		"repo",
		SoftwareRepoPath,
		"Path of the software repository.",
	)
	cmdOptions.removeCmd.StringVar(
		&cmdOptions.softwareType,
		"type",
		"",
		"Type of the software.",
	)
	cmdOptions.removeCmd.StringVar(
		&cmdOptions.softwareName,
		"filename",
		"",
		"File name of the software.",
	)
}

// Remove the specified software package from the software repo.
func Remove(swName, swType, swRepo string) error {
	log.Printf("Entering repo::Remove(%s, %s, %s)", swName, swType, swRepo)
	defer log.Println("Exiting repo::Remove")

	if swRepo == "" {
		return logutil.PrintNLogError("Unable to remove %s software %s. "+
			"Failed to determine software repository.",
			swType, swName)
	}
	if swName != "" && swType == "" {
		return logutil.PrintNLogError("Invalid usage. Software type must be specified when software name is specified.")
	}

	absSwPath := filepath.Clean(filepath.FromSlash(swRepo +
		string(os.PathSeparator) + strings.ToLower(swType) + string(os.PathSeparator) +
		swName))

	fi, err := os.Stat(absSwPath)
	if err != nil {
		log.Printf("Unable to stat on %s: %+v. Error: %s\n",
			absSwPath, fi, err.Error())
		return logutil.PrintNLogError("Unable to remove %s software %s. "+
			"Specified software not found.",
			swType, swName)
	}

	err = osutils.OsRemoveAll(absSwPath)
	if err != nil {
		log.Printf("Unable to remove on %s. Error: %s\n",
			absSwPath, err.Error())
		return logutil.PrintNLogError("Failed to remove %s software %s.",
			swType, swName)
	}

	log.Printf("Successfully removed %s software", absSwPath)
	logutil.PrintNLog("Successfully removed %s software %s from repository.\n",
		swType, swName)
	return nil
}
