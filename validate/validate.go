// Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package validate

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	logger "github.com/VeritasOS/plugin-manager/utils/log"
	"github.com/VeritasOS/software-update-manager/repo"
	"github.com/VeritasOS/software-update-manager/validate/version"
)

var execCommand = exec.Command

const syscfgCmd = "/usr/bin/syscfg/syscfg"

func isRPMCompatibile(productVersion string, rpmFile string) error {
	logger.Debug.Printf("Entering validate::isRPMCompatibile(%s, %s)...",
		productVersion, rpmFile)
	defer logger.Debug.Println("Exiting validate::isRPMCompatibile")

	err := fileExists(rpmFile)
	if err != nil {
		return err
	}

	info, err := repo.ListRPMFilesInfo([]string{rpmFile}, productVersion)
	if err != nil {
		return err
	}

	curInfo := info[0]
	if !version.Compare(productVersion, curInfo.GetMatchedVersion()) {
		return logger.ConsoleError.PrintNReturnError("The %s software file is not compatibile for %s version.",
			filepath.Base(rpmFile), productVersion)
	}
	return nil
}

func isSigned(rpmFile string) error {
	signCommand := "rpm -qip '" + rpmFile + "' | grep Signature | cut -d ':' -f 2-"
	rpmCmd := execCommand("sh", "-c", signCommand)
	out, err := rpmCmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("Failed to execute RPM command "+signCommand+
			". Error %v", err.Error())
		return err
	}

	rpmSign := strings.TrimSpace(string(out))
	if strings.Contains(rpmSign, "(none)") {
		return logger.ConsoleError.PrintNReturnError("RPM file %s is not signed. Only "+
			"install updates that have been downloaded from or provided by Veritas",
			rpmFile)
	}

	return nil
}

func verifySign(rpmFile string) error {
	/*
	 * In RHEL8, RPM installation fails if the RPM is built without digest.
	 * Set the separation build time as: Dec 05, 2022 00:00:00 GMT (Gordon GA date)(1670198400)
	 * before that build time, the rpms are built without package or header digests, skip digest check
	 * after that build time, the rpms are built with digest, will use normal rpm installation.
	 */
	var rpmCmd *exec.Cmd
	cmdString := "/bin/rpm -qp --queryformat '%{BUILDTIME}' '" + rpmFile + "'"
	rpmCmd = execCommand("/bin/sh", "-c", cmdString)
	out, err := rpmCmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("Failed to execute RPM command "+cmdString+". Error %v", err.Error())
		return err
	}
	if string(out) < "1670198400" {
		rpmCmd = execCommand("/bin/rpm", "-Kv", "--nodigest", rpmFile)
	} else {
		rpmCmd = execCommand("/bin/rpm", "-Kv", rpmFile)
	}
	out, err = rpmCmd.CombinedOutput()
	logger.Debug.Println(string(out))

	if err != nil {
		err = logger.ConsoleError.PrintNReturnError("Signature validation failed for %s. "+
			"Only install updates that have been downloaded from or provided by "+
			"Veritas. Error %v",
			rpmFile, err.Error())
		return err
	}
	return nil
}

func validateSignature(rpmFile string) error {

	if err := fileExists(rpmFile); err != nil {
		return err
	}

	if err := isSigned(rpmFile); err != nil {
		return err
	}

	if err := verifySign(rpmFile); err != nil {
		return err
	}

	return nil
}

func fileExists(filePath string) error {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		err := fmt.Errorf(filePath + " file does not exist")
		return err
	}
	return nil
}

func validateBiosPassword(biosPassword string) error {
	if _, err := os.Stat("/sys/firmware/efi"); os.IsNotExist(err) {
		// Not an EFI system
		return nil
	}
	if _, err := os.Stat(syscfgCmd); os.IsNotExist(err) {
		// syscfg command not found
		return fmt.Errorf("syscfg command not found")
	}

	var validateCmd *exec.Cmd
	vCommand := syscfgCmd + " /bcs '" + biosPassword + "' 'Advanced' 'Intel(R) Virtualization Technology' 1"
	validateCmd = execCommand("sh", "-c", vCommand)

	_, err := validateCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("bios password is incorrect. Error %v", err.Error())
	}
	return nil
}

// Exec executes rpm validation
func Exec(args []string) {
	os.Args = args

	versionFlag := flag.Bool("version", false, "command for validating rpm version: validate -version -rpm=<path> -product-version=<version> ")
	signFlag := flag.Bool("signature", false, "command for validating rpm version: validate -signature -rpm=<path> ")
	biosPwFlag := flag.Bool(
		"biospassword",
		false,
		`Validate BIOS password.
Set environment variable 'BIOS_PASSWORD' to validate password.
Using env var to avoid password getting logged and also process list (ps) showing password`,
	)
	rpmFile := flag.String("rpm", "", "path of rpm file")
	productVersion := flag.String("product-version", "", "product version")

	flag.Parse()

	if *versionFlag {
		if *productVersion == "" || *rpmFile == "" {
			flag.PrintDefaults()
			os.Exit(1)
		}
		err := isRPMCompatibile(*productVersion, *rpmFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
			os.Exit(1)
		}
	}
	if *signFlag {
		if *rpmFile == "" {
			flag.PrintDefaults()
			os.Exit(1)
		}
		err := validateSignature(*rpmFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
			os.Exit(1)
		}
	}

	if *biosPwFlag {
		// INFO: Using environment variable BIOS_PASSWORD to validate password, so as to avoid
		// 1. password getting logged
		// 2. process list (ps) showing password
		passwd := os.Getenv("BIOS_PASSWORD")
		err := validateBiosPassword(passwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
			os.Exit(1)
		}
	}

	if !*signFlag && !*versionFlag && !*biosPwFlag {
		flag.PrintDefaults()
		os.Exit(1)
	}
}
