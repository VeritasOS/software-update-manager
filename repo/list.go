// Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package repo defines software repository functions like listing, removing
// packages from software repository.
package repo

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	logger "github.com/VeritasOS/plugin-manager/utils/log"
	"github.com/VeritasOS/plugin-manager/utils/output"
	"github.com/VeritasOS/software-update-manager/utils/rpm"
	"github.com/VeritasOS/software-update-manager/validate/version"

	"gopkg.in/yaml.v3"
)

// FormatVersionName is the ASUM RPM format version string that's embedded
// into RPM used for identifying the JSON format version.
const FormatVersionName = "ASUM RPM Format Version"

var (
	fGetRPMPackageInfo = rpm.GetRPMPackageInfo
	fGetChecksum       = getChecksum
	fListRepo          = listRepo
	fListRPMFilesInfo  = ListRPMFilesInfo
	fOpen              = os.Open
)

// RPMInfo is the list of RPM package info
type RPMInfo interface {
	GetMatchedVersion() string
	GetRPMName() string
	GetRPMRelease() string
	GetRPMType() string
	GetRPMVersion() string
}

// Version 2 RPM Information related fields & helper functions below:

// v2productVersion is the details for a given product-version from the
// version compatibility matrix/info JSON.
type v2productVersion struct {
	Install struct {
		ConfirmationMessage []string `yaml:"confirmation-message"`
		EstimatedMinutes    uint     `yaml:"estimated-minutes"`
		RequiresRestart     bool     `yaml:"requires-restart"`
		SupportsPrecheck    bool     `yaml:"supports-precheck"`
		SupportsRollback    bool     `yaml:"supports-rollback"`
	} `yaml:",omitempty"`
	Rollback struct {
		ConfirmationMessage []string `yaml:"confirmation-message"`
		EstimatedMinutes    uint     `yaml:"estimated-minutes"`
		RequiresRestart     bool     `yaml:"requires-restart"`
	} `yaml:",omitempty"`
	Commit struct {
		ConfirmationMessage []string `yaml:"confirmation-message"`
		EstimatedMinutes    uint     `yaml:"estimated-minutes"`
	} `yaml:",omitempty"`
}

// v2RPMInfo is the list of RPM package info
type v2RPMInfo struct {
	// Name of the RPM
	Name string
	// RPM file name
	FileName         string
	Description      []string
	Type             string
	URL              string
	Version          string
	Release          string
	Checksum         string `yaml:"checksum,omitempty"`
	DisplayType      string `yaml:"display-type,omitempty"`
	BuildDate        string
	matchedVersion   string
	v2productVersion `yaml:",inline"`
}

// GetRPMName returns the name of the RPM.
func (v2 v2RPMInfo) GetRPMName() string {
	return v2.Name
}

// GetRPMRelease returns the release number of the RPM.
func (v2 v2RPMInfo) GetRPMRelease() string {
	return v2.Release
}

// GetRPMType returns the type of the RPM.
func (v2 v2RPMInfo) GetRPMType() string {
	return v2.Type
}

// GetRPMVersion returns the release number of the RPM.
func (v2 v2RPMInfo) GetRPMVersion() string {
	return v2.Version
}

// GetMatchedVersion retrives the supported product-version from
// version compatibility matrix.
func (v2 v2RPMInfo) GetMatchedVersion() string {
	return v2.matchedVersion
}

// Version 1 RPM Information related fields & helper functions below:

// v1RPMInfo is the list of RPM package info
type v1RPMInfo struct {
	Description []string `yaml:"description"`
	Estimate    struct {
		Hours   string `yaml:"hours"`
		Minutes string `yaml:"minutes"`
		Seconds string `yaml:"seconds"`
	} `yaml:"estimate"`
	// RPM file name
	FileName       string `yaml:"filename"`
	Name           string `yaml:"name"`
	matchedVersion string
	Reboot         string `yaml:"reboot"`
	Summary        string `yaml:"summary"`
	Type           string `yaml:"type"`
	URL            string `yaml:"url"`
	Version        string `yaml:"version"`
	Checksum       string `yaml:"checksum,omitempty"`
	Release        string `yaml:"release"`
}

// GetRPMName returns the name of the RPM.
func (v1 v1RPMInfo) GetRPMName() string {
	return v1.Name
}

// GetRPMRelease returns the release number of the RPM.
func (v1 v1RPMInfo) GetRPMRelease() string {
	return v1.Release
}

// GetRPMType returns the type of the RPM.
func (v1 v1RPMInfo) GetRPMType() string {
	return v1.Type
}

// GetRPMVersion returns the release number of the RPM.
func (v1 v1RPMInfo) GetRPMVersion() string {
	return v1.Version
}

// GetMatchedVersion retrives the supported product-version from
// version compatibility matrix.
func (v1 v1RPMInfo) GetMatchedVersion() string {
	return v1.matchedVersion
}

// List the packages present in the software repo along with their details.
func List(params map[string]string) ([]RPMInfo, error) {
	logger.Debug.Printf("Entering repo::List(%v)", params)
	defer logger.Debug.Println("Exiting repo::List")

	productVersion := params["productVersion"]

	var info []RPMInfo

	files, err := fListRepo(params)
	if err != nil {
		return info, err
	}
	includeFields := params["includeFields"]
	m := strings.Split(includeFields, ",")

	info, err = fListRPMFilesInfo(files, productVersion, m...)
	if err != nil {
		return info, err
	}

	output.Write(info)

	return info, nil
}

func listRepo(params map[string]string) ([]string, error) {
	logger.Debug.Printf("Entering repo::listRepo(%v)", params)
	defer logger.Debug.Println("Exiting repo::listRepo")

	swName := params["softwareName"]
	swRepo := params["softwareRepo"]
	swType := strings.ToLower(params["softwareType"])

	var files []string
	swTypes := []string{}

	// If software type is not specified, then get the list of packages for
	// 	all types.
	if "" != swType {
		swTypes = append(swTypes, swType)
	} else {
		if _, err := os.Stat(swRepo); os.IsNotExist(err) {
			logger.ConsoleWarning.Printf("Software repository '%s' does not exist.", swRepo)
			return files, nil
		}
		dirs, err := ioutil.ReadDir(swRepo)
		if err != nil {
			logger.Error.Printf("ioutil.ReadDir(%s); err=%s", swRepo, err.Error())
			return files, logger.ConsoleError.PrintNReturnError("Failed to get contents of software repository.")
		}

		for _, dir := range dirs {
			curDir := filepath.FromSlash(swRepo + "/" + dir.Name())
			fi, err := os.Stat(curDir)
			if err != nil {
				logger.Error.Printf("Unable to stat on %s directory. err=%s", dir, err.Error())
				continue
			}
			if !fi.IsDir() {
				logger.Error.Printf("%s is not a directory.", curDir)
				continue
			}

			swTypes = append(swTypes, dir.Name())
		}
	}
	for _, dir := range swTypes {
		curDir := filepath.Clean(filepath.FromSlash(swRepo +
			string(os.PathSeparator) + dir))

		tfiles, err := ioutil.ReadDir(curDir)
		if err != nil {
			logger.Error.Printf("Unable to read contents of %s directory. err=%s\n",
				curDir, err.Error())
		}
		logger.Debug.Printf("%s files: %v", dir, tfiles)
		for _, tf := range tfiles {
			logger.Debug.Printf("Package: %v", tf)
			matched, err := regexp.MatchString("[.]rpm$", tf.Name())
			if err != nil {
				logger.Error.Printf("regexp.MatchString(%s, %s); err=%s",
					"[.]rpm", tf.Name(), err.Error())
				continue
			}
			// If not an RPM file, skip
			if !matched {
				continue
			}

			if "" == swName || tf.Name() == swName {
				files = append(files, filepath.FromSlash(curDir+
					string(os.PathSeparator)+tf.Name()))
			}
		}
	}

	return files, nil
}

func getChecksum(src io.Reader) string {
	h := sha256.New()
	if _, err := io.Copy(h, src); err != nil {
		logger.ConsoleError.PrintNReturnError("Failed to get RPM file checksum. Error: [%v]", err.Error())
	}
	return hex.EncodeToString(h.Sum(nil))

}

// ListRPMFilesInfo lists the info of the RPM files.
// includeFields is a comma separated values you can provide to include extra fields other than default fields.
// As of now only "checksum" is extra field you can populate.
// If you are adding new field please make sure it should not be populated by default
// you need to use --include-fields flag. Other wise
// if you add new field in response then response time of ListRPMFilesInfo function may increase
func ListRPMFilesInfo(files []string, productVersion string, includeFields ...string) ([]RPMInfo, error) {
	logger.Debug.Printf("Entering repo::ListRPMFilesInfo(%v, %v)", files, productVersion)
	defer logger.Debug.Println("Exiting repo::ListRPMFilesInfo")

	var info []RPMInfo
	for _, file := range files {
		metaData, err := fGetRPMPackageInfo(filepath.FromSlash(file))
		if err != nil {
			return info, logger.ConsoleError.PrintNReturnError("Failed to get software details.")
		}
		parsedData := rpm.ParseMetaData(string(metaData))

		rpmName, err := rpm.GetRPMPackageName(filepath.FromSlash(file))
		if err != nil {
			logger.ConsoleError.PrintNReturnError("Failed to get software name.")
			rpmName = filepath.Base(file)
		}

		var checksum string
		for _, v := range includeFields {
			if v == "checksum" {
				f, err := fOpen(file)
				if err != nil {
					logger.ConsoleError.PrintNReturnError("Failed to get RPM file checksum. Error: [%v]", err.Error())
				}
				checksum = fGetChecksum(f)

				defer f.Close()
			}
		}

		if _, ok := parsedData[FormatVersionName]; !ok {
			listData := v1RPMInfo{
				Description: []string{parsedData["Description"]},
				FileName:    filepath.Base(file),
				Name:        rpmName,
				Summary:     parsedData["Summary"],
				Type:        parsedData["Type"],
				URL:         parsedData["URL"],
				Version:     parsedData["Version"],
				Reboot:      "n/a",
				Release:     parsedData["Release"],
			}

			listData.Estimate.Hours = "0"
			listData.Estimate.Minutes = "0"
			listData.Estimate.Seconds = "0"

			listData.Checksum = checksum
			if "" != productVersion {
				versionInfo, err := version.GetCompatibileVersionInfo(productVersion, parsedData["VersionInfo"])
				//In case of error, i.e the version of rpm and product version is not compatible we ignore error and contiune execution
				//This error is expected when the rpm is already applied. List API should still return rpm details
				if err != nil {
					logger.Error.Printf("Error in GetCompatibileVersionInfo, err=%v", err)
				}

				listData.Description = append(listData.Description, versionInfo.Description...)
				listData.matchedVersion = versionInfo.Version
				listData.Reboot = versionInfo.Reboot
				listData.Estimate.Hours = versionInfo.Estimate.Hours
				listData.Estimate.Minutes = versionInfo.Estimate.Minutes
				listData.Estimate.Seconds = versionInfo.Estimate.Seconds
			}

			info = append(info, listData)
		} else {
			logger.Debug.Printf("%s: %v", FormatVersionName, parsedData[FormatVersionName])

			listData := v2RPMInfo{
				FileName: filepath.Base(file),
				Name:     parsedData["Name"],
				Version:  parsedData["Version"],
				Release:  parsedData["Release"],
				URL:      parsedData["URL"],
			}
			rpmInfo := parsedData["RPM Info"]
			err := yaml.Unmarshal([]byte(rpmInfo), &listData)
			if err != nil {
				logger.Error.Printf("yaml.Unmarshal(%s, %+v); err=%s",
					rpmInfo, &listData, err.Error())
			}
			listData.BuildDate = parsedData["Build Date"]
			listData.Checksum = checksum
			if "" != productVersion {
				allVersionsInfo := struct {
					VersionInfo []struct {
						Version          string `yaml:"product-version"`
						v2productVersion `yaml:",inline"`
						Description      []string `yaml:"description"`
					} `yaml:"compatibility-info"`
				}{}

				err := yaml.Unmarshal([]byte(rpmInfo), &allVersionsInfo)
				if err != nil {
					logger.Error.Printf("yaml.Unmarshal(%s, %+v); err=%s",
						rpmInfo, &allVersionsInfo, err.Error())
				}
				// INFO: First check Version as-is,
				// 	if there is no match, then do pattern comparison.
				for _, vInfo := range allVersionsInfo.VersionInfo {
					if productVersion == vInfo.Version {
						listData.v2productVersion = vInfo.v2productVersion
						listData.matchedVersion = vInfo.Version
						listData.Description = append(listData.Description, vInfo.Description...)
					}
				}
				if listData.matchedVersion == "" {
					for _, vInfo := range allVersionsInfo.VersionInfo {
						if version.Compare(productVersion, vInfo.Version) {
							listData.v2productVersion = vInfo.v2productVersion
							listData.matchedVersion = vInfo.Version
						}
					}
				}
			}

			info = append(info, listData)
		}
	}

	return info, nil
}

// registerCommandList registers the list command that enables one to
// view the RPMs present in the software update repository.
func registerCommandList(progname string) {
	logger.Debug.Printf("Entering repo::registerCommandList(%s)", progname)
	defer logger.Debug.Println("Exiting repo::registerCommandList")

	cmdOptions.listCmd = flag.NewFlagSet(progname+" list", flag.PanicOnError)

	cmdOptions.listCmd.StringVar(
		&cmdOptions.productVersion,
		"product-version",
		"",
		"Version that a software should be compatibile with."+
			" (I.e., product-version)",
	)
	cmdOptions.listCmd.StringVar(
		&cmdOptions.softwareName,
		"filename",
		"",
		"File name of the software.",
	)
	cmdOptions.listCmd.StringVar(
		&cmdOptions.softwareRepo,
		"repo",
		SoftwareRepoPath,
		"Path of the software repository.",
	)
	cmdOptions.listCmd.StringVar(
		&cmdOptions.softwareType,
		"type",
		"",
		"Type of the software.",
	)
	cmdOptions.listCmd.StringVar(
		&cmdOptions.includeFields,
		"include-fields",
		"",
		"In -include-fields comma separated values can be define. "+
			"By default checksum field is not populate in response."+
			"By using -include-fields='checksum' flag it will get populated in response",
	)
	output.RegisterCommandOptions(cmdOptions.listCmd,
		map[string]string{"output-format": "yaml"})
}
