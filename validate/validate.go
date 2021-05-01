// Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package validate

import (
	"flag"
	"fmt"
	logutil "github.com/VeritasOS/plugin-manager/utils/log"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"software-update-manager/repo"
	"software-update-manager/validate/version"
	"strings"
)

var execCommand = exec.Command

func isRPMCompatibile(productVersion string, rpmFile string) error {
	log.Printf("Entering validate::isRPMCompatibile(%s, %s)...",
		productVersion, rpmFile)
	defer log.Println("Exiting validate::isRPMCompatibile")

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
		return logutil.PrintNLogError("The %s software file is not compatibile for %s version.",
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
		return logutil.PrintNLogError("RPM file %s is not signed. Only "+
			"install updates that have been downloaded from or provided by Veritas",
			rpmFile)
	}

	return nil
}

func verifySign(rpmFile string) error {

	rpmCmd := execCommand("rpm", "-Kv", rpmFile)
	out, err := rpmCmd.CombinedOutput()
	log.Printf(string(out))

	if err != nil {
		err = logutil.PrintNLogError("Signature validation failed for %s. "+
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

// Exec executes rpm validation
func Exec(args []string) {
	os.Args = args

	versionFlag := flag.Bool("version", false, "command for validating rpm version: validate -version -rpm=<path> -product-version=<version> ")
	signFlag := flag.Bool("signature", false, "command for validating rpm version: validate -signature -rpm=<path> ")
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
	if !*signFlag && !*versionFlag {
		flag.PrintDefaults()
		os.Exit(1)
	}
}
