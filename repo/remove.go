// Copyright (c) 2022 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package repo defines software repository functions like listing, removing
//
//	packages from software repository.
package repo

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	logger "github.com/VeritasOS/plugin-manager/utils/log"
	osutils "github.com/VeritasOS/plugin-manager/utils/os"
)

// registerCommandRemove registers the remove command that enables one to
//
//	remove the RPM of the specified type from the software update repository.
func registerCommandRemove(progname string) {
	logger.Debug.Printf("Entering repo::registerCommandRemove(%s)", progname)
	defer logger.Debug.Println("Exiting repo::registerCommandRemove")

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
	logger.Debug.Printf("Entering repo::Remove(%s, %s, %s)", swName, swType, swRepo)
	defer logger.Debug.Println("Exiting repo::Remove")

	if swRepo == "" {
		return logger.ConsoleError.PrintNReturnError("Unable to remove %s software %s. "+
			"Failed to determine software repository.",
			swType, swName)
	}
	if swName != "" && swType == "" {
		return logger.ConsoleError.PrintNReturnError("Invalid usage. Software type must be specified when software name is specified.")
	}

	absSwPath := filepath.Clean(filepath.FromSlash(swRepo +
		string(os.PathSeparator) + strings.ToLower(swType) + string(os.PathSeparator) +
		swName))

	fi, err := os.Stat(absSwPath)
	if err != nil {
		logger.Error.Printf("Unable to stat on %s: %+v, err=%s", absSwPath, fi, err.Error())
		return logger.ConsoleError.PrintNReturnError("Unable to remove %s software %s. "+
			"Specified software not found.",
			swType, swName)
	}

	err = osutils.OsRemoveAll(absSwPath)
	if err != nil {
		logger.Error.Printf("Unable to remove on %s, err=%s", absSwPath, err.Error())
		return logger.ConsoleError.PrintNReturnError("Failed to remove %s software %s.",
			swType, swName)
	}

	logger.Info.Printf("Successfully removed %s software", absSwPath)
	logger.ConsoleInfo.Printf("Successfully removed %s software %s from repository.",
		swType, swName)
	return nil
}
