// Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package rpm contains utility functions (required by SUM) for managing RPM files.
package rpm

import (
	logutil "github.com/VeritasOS/plugin-manager/utils/log"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Cmd is the path of the `rpm` command.
const Cmd = "/usr/bin/rpm"

// GetRPMPackageInfo queries and retrieves RPM file info.
func GetRPMPackageInfo(rpmPath string) ([]byte, error) {
	log.Printf("Entering rpm::GetRPMPackageInfo(%s)", rpmPath)
	defer log.Println("Exiting rpm::GetRPMPackageInfo")

	cmdParams := []string{"-q", "-p", "--info", filepath.FromSlash(rpmPath)}
	cmd := exec.Command(os.ExpandEnv(Cmd), cmdParams...)
	stdOutErr, err := cmd.CombinedOutput()
	log.Println("Stdout & Stderr:", string(stdOutErr))
	if err != nil {
		log.Printf("Failed to get %s RPM details. Error: %s\n",
			rpmPath, err.Error())
		return stdOutErr, err
	}
	return stdOutErr, nil
}

// IsInstalled tells whether RPM is installed on the system.
func IsInstalled(rpmName string) bool {
	log.Printf("Entering rpm::IsInstalled(%s)", rpmName)
	defer log.Println("Exiting rpm::IsInstalled")

	cmdParams := []string{"-q", rpmName}
	cmd := exec.Command(os.ExpandEnv(Cmd), cmdParams...)
	stdOutErr, err := cmd.CombinedOutput()
	log.Println("Stdout & Stderr:", string(stdOutErr))
	if err != nil {
		log.Printf("Failed to query on %s RPM. Error: %s\n",
			rpmName, err.Error())
		return false
	}
	return true
}

// Install the specified RPM file.
func Install(rpmPath string) error {
	log.Printf("Entering rpm::Install(%s)", rpmPath)
	defer log.Println("Exiting rpm::Install")

	cmdParams := []string{"-Uvh", filepath.FromSlash(rpmPath)}
	cmd := exec.Command(os.ExpandEnv(Cmd), cmdParams...)
	stdOutErr, err := cmd.CombinedOutput()
	log.Println("Stdout & Stderr:", string(stdOutErr))
	if err != nil {
		log.Printf("Failed to install %s RPM. Error: %s\n",
			rpmPath, err.Error())
		return err
	}
	return nil
}

// ParseMetaData parses the RPM metadata
// 	into key-value pair.
func ParseMetaData(metaData string) map[string]string {
	log.Printf("Entering rpm::ParseMetaData()")
	defer log.Println("Exiting rpm::ParseMetaData")

	parsedData := map[string]string{}
	key := ""
	// INFO: pattern matches "<key name> : <value field>" of the RPM metadata.
	pattern := `^\w+(\s*\w*)*\s*:`
	for _, line := range strings.Split(metaData, "\n") {
		matched, err := regexp.MatchString(pattern, line)
		if err != nil {
			log.Printf("regexp.MatchString(%s, %s); Error: %s",
				"[.]rpm", line, err.Error())
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
	log.Printf("Entering rpm::Uninstall(%s)", rpmName)
	defer log.Println("Exiting rpm::Uninstall")

	cmdParams := []string{"-e", rpmName}
	cmd := exec.Command(os.ExpandEnv(Cmd), cmdParams...)
	stdOutErr, err := cmd.CombinedOutput()
	log.Println("Stdout & Stderr:", string(stdOutErr))
	if err != nil {
		log.Printf("Failed to remove %s RPM. Error: %s\n",
			rpmName, err.Error())
		return logutil.PrintNLogError("Failed to uninstall %s software.", rpmName)
	}
	return nil
}
