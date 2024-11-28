// Copyright (c) 2022 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package rpm contains utility functions (required by ASUM) for managing RPM files.
package rpm

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	logger "github.com/VeritasOS/plugin-manager/utils/log"
)

// Cmd is the path of the `rpm` command.
const Cmd = "/usr/bin/rpm"

// GetRPMPackageInfo queries and retrieves RPM file info.
func GetRPMPackageInfo(rpmPath string) ([]byte, error) {
	logger.Debug.Printf("Entering rpm::GetRPMPackageInfo(%s)", rpmPath)
	defer logger.Info.Println("Exiting rpm::GetRPMPackageInfo")

	cmdParams := []string{"-q", "-p", "--info", filepath.FromSlash(rpmPath)}
	cmd := exec.Command(os.ExpandEnv(Cmd), cmdParams...)
	stdOutErr, err := cmd.CombinedOutput()
	logger.Debug.Println("Stdout & Stderr:", string(stdOutErr))
	if err != nil {
		logger.Error.Printf("Failed to get %s RPM details, err=%s",
			rpmPath, err.Error())
		return stdOutErr, err
	}
	return stdOutErr, nil
}

// GetRPMPackageName queries and retrieves RPM file name.
func GetRPMPackageName(rpmPath string) (string, error) {
	logger.Debug.Printf("Entering rpm::GetRPMPackageName(%s)", rpmPath)
	defer logger.Debug.Println("Exiting rpm::GetRPMPackageName")

	cmdParams := []string{"-q", "-p", filepath.FromSlash(rpmPath)}
	cmd := exec.Command(os.ExpandEnv(Cmd), cmdParams...)
	stdOutErr, err := cmd.CombinedOutput()
	logger.Debug.Println("Stdout & Stderr:", string(stdOutErr))
	if err != nil {
		logger.Error.Printf("Failed to get %s RPM details, err=%s",
			rpmPath, err.Error())
		return string(stdOutErr), err
	}
	rpmName := string(stdOutErr)
	rpmName = strings.TrimSuffix(rpmName, "\n")
	return rpmName, nil

}

// IsInstalled tells whether RPM is installed on the system.
func IsInstalled(rpmName string) bool {
	logger.Debug.Printf("Entering rpm::IsInstalled(%s)", rpmName)
	defer logger.Debug.Println("Exiting rpm::IsInstalled")

	cmdParams := []string{"-q", rpmName}
	cmd := exec.Command(os.ExpandEnv(Cmd), cmdParams...)
	stdOutErr, err := cmd.CombinedOutput()
	logger.Debug.Println("Stdout & Stderr:", string(stdOutErr))
	if err != nil {
		logger.Error.Printf("Failed to query on %s RPM, err=%s",
			rpmName, err.Error())
		return false
	}
	return true
}

// Install the specified RPM file.
func Install(rpmPath string) error {
	logger.Debug.Printf("Entering rpm::Install(%s)", rpmPath)
	defer logger.Debug.Println("Exiting rpm::Install")

	cmdParams := []string{"-Uvh", filepath.FromSlash(rpmPath)}
	cmd := exec.Command(os.ExpandEnv(Cmd), cmdParams...)
	stdOutErr, err := cmd.CombinedOutput()
	logger.Debug.Println("Stdout & Stderr:", string(stdOutErr))
	if err != nil {
		logger.Error.Printf("Failed to install %s RPM, err=%s",
			rpmPath, err.Error())
		return err
	}
	return nil
}

// ParseMetaData parses the RPM metadata
//
//	into key-value pair.
func ParseMetaData(metaData string) map[string]string {
	logger.Debug.Printf("Entering rpm::ParseMetaData()")
	defer logger.Debug.Println("Exiting rpm::ParseMetaData")

	parsedData := map[string]string{}
	key := ""
	// INFO: pattern matches "<key name> : <value field>" of the RPM metadata.
	pattern := `^\w+(\s*\w*)*\s*:`
	for _, line := range strings.Split(metaData, "\n") {
		matched, err := regexp.MatchString(pattern, line)
		if err != nil {
			logger.Error.Printf("regexp.MatchString(%s, %s), err=%s", pattern, line, err.Error())
			continue
		}
		if matched {
			fields := strings.Split(line, ":")
			key = strings.TrimSpace(fields[0])
			if len(fields) > 1 {
				parsedData[key] = strings.TrimSpace(
					strings.Join(fields[1:], ":"))
			}
		} else {
			parsedData[key] += strings.TrimSpace(line)
		}
	}
	return parsedData
}

// Uninstall erases/uninstalls the specified RPM from node.
func Uninstall(rpmName string) error {
	logger.Debug.Printf("Entering rpm::Uninstall(%s)", rpmName)
	defer logger.Debug.Println("Exiting rpm::Uninstall")

	cmdParams := []string{"-e", rpmName}
	cmd := exec.Command(os.ExpandEnv(Cmd), cmdParams...)
	stdOutErr, err := cmd.CombinedOutput()
	logger.Debug.Println("Stdout & Stderr:", string(stdOutErr))
	if err != nil {
		logger.Error.Printf("Failed to remove %s RPM, err=%s", rpmName, err.Error())
		return logger.ConsoleError.PrintNReturnError("Failed to uninstall %s software.", rpmName)
	}
	return nil
}
