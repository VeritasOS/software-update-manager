// Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package repo defines software repository functions like listing, removing
// 	packages from software repository.
package repo

import (
	"flag"
	logutil "github.com/VeritasOS/plugin-manager/utils/log"
	"github.com/VeritasOS/plugin-manager/utils/output"
	"github.com/VeritasOS/software-update-manager/utils/rpm"
	"github.com/VeritasOS/software-update-manager/validate/version"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// FormatVersionName is the ASUM RPM format version string that's embedded
// 	into RPM used for identifying the JSON format version.
const FormatVersionName = "ASUM RPM Format Version"

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
// 	version compatibility matrix/info JSON.
type v2productVersion struct {
	Install struct {
		ConfirmationMessage []string `yaml:"confirmation-message"`
		EstimatedMinutes    uint     `yaml:"estimated-minutes"`
		RequiresRestart     bool     `yaml:"requires-restart"`
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
	FileName    string
	Description []string
	Type        string
	URL         string
	Version     string
	Release     string
	// NOTE: The time.Time value is getting chopped off while dumping output
	// 	in json at ansible layer causing json unmarshal failure at consumer.
	// 	so commenting for now.
	// BuildDate   time.Time

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
// 	version compatibility matrix.
func (v2 v2RPMInfo) GetMatchedVersion() string {
	return v2.matchedVersion
}

// Version 1 RPM Information related fields & helper functions below:

// v1RPMInfo is the list of RPM package info
type v1RPMInfo struct {
	Description string `yaml:"description"`
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
}

// GetRPMName returns the name of the RPM.
func (v1 v1RPMInfo) GetRPMName() string {
	return v1.Name
}

// GetRPMRelease returns the release number of the RPM.
func (v1 v1RPMInfo) GetRPMRelease() string {
	// v1 doesn't support displaying release number, so return empty.
	return ""
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
// 	version compatibility matrix.
func (v1 v1RPMInfo) GetMatchedVersion() string {
	return v1.matchedVersion
}

func parseDate(rawDate string) (time.Time, error) {
	const dateLayout = "Mon 02 Jan 2006 03:04:05 PM MST"
	t, err := time.Parse(dateLayout, rawDate)
	if err != nil {
		// INFO: On few systems, the build date appeared in
		// 	ANSIC layout, so, try parsing with that.
		t, err = time.Parse(time.ANSIC, rawDate)
	}
	if err != nil {
		log.Printf("Failed to parse date: %s. Error: %s\n", rawDate, err)
	}
	return t, err
}

// List the packages present in the software repo along with their details.
func List(params map[string]string) ([]RPMInfo, error) {
	log.Printf("Entering repo::List(%v)", params)
	defer log.Println("Exiting repo::List")

	productVersion := params["productVersion"]

	var info []RPMInfo

	files, err := listRepo(params)
	if err != nil {
		return info, err
	}

	info, err = ListRPMFilesInfo(files, productVersion)
	if err != nil {
		return info, err
	}

	output.Write(info)

	return info, nil
}

func listRepo(params map[string]string) ([]string, error) {
	log.Printf("Entering repo::listRepo(%v)", params)
	defer log.Println("Exiting repo::listRepo")

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
			logutil.PrintNLogWarning("Software repository '%s' does not exist.",
				swRepo)
			return files, nil
		}
		dirs, err := ioutil.ReadDir(swRepo)
		if err != nil {
			log.Printf("ioutil.ReadDir(%s); Error: %s", swRepo, err.Error())
			return files, logutil.PrintNLogError("Failed to get contents of software repository.")
		}

		for _, dir := range dirs {
			curDir := filepath.FromSlash(swRepo + "/" + dir.Name())
			fi, err := os.Stat(curDir)
			if err != nil {
				log.Printf("Unable to stat on %s directory. Error: %s\n",
					dir, err.Error())
				continue
			}
			if !fi.IsDir() {
				log.Printf("%s is not a directory.\n", curDir)
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
			log.Printf("Unable to read contents of %s directory. Error: %s\n",
				curDir, err.Error())
		}
		log.Printf("%s files: %v", dir, tfiles)
		for _, tf := range tfiles {
			log.Printf("Package: %v", tf)
			matched, err := regexp.MatchString("[.]rpm$", tf.Name())
			if err != nil {
				log.Printf("regexp.MatchString(%s, %s); Error: %s",
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

// ListRPMFilesInfo lists the info of the RPM files.
func ListRPMFilesInfo(files []string, productVersion string) ([]RPMInfo, error) {
	log.Printf("Entering repo::ListRPMFilesInfo(%v, %v)", files, productVersion)
	defer log.Println("Exiting repo::ListRPMFilesInfo")

	var info []RPMInfo
	for _, file := range files {
		metaData, err := rpm.GetRPMPackageInfo(filepath.FromSlash(file))
		if err != nil {
			return info, logutil.PrintNLogError("Failed to get software details.")
		}
		parsedData := rpm.ParseMetaData(string(metaData))

		if _, ok := parsedData[FormatVersionName]; !ok {
			listData := v1RPMInfo{
				Description: parsedData["Description"],
				FileName:    filepath.Base(file),
				Name:        filepath.Base(file),
				Summary:     parsedData["Summary"],
				Type:        parsedData["Type"],
				URL:         parsedData["URL"],
				Version:     parsedData["Version"],
				Reboot:      "n/a",
			}
			listData.Estimate.Hours = "0"
			listData.Estimate.Minutes = "0"
			listData.Estimate.Seconds = "0"
			if "" != productVersion {
				versionInfo, err := version.GetCompatibileVersionInfo(productVersion, parsedData["VersionInfo"])
				//In case of error, i.e the version of rpm and product version is not compatible we ignore error and contiune execution
				//This error is expected when the rpm is already applied. List API should still return rpm details
				if err != nil {
					log.Printf("Error in GetCompatibileVersionInfo::(%s)...",
						err)
				}
				listData.matchedVersion = versionInfo.Version
				listData.Reboot = versionInfo.Reboot
				listData.Estimate.Hours = versionInfo.Estimate.Hours
				listData.Estimate.Minutes = versionInfo.Estimate.Minutes
				listData.Estimate.Seconds = versionInfo.Estimate.Seconds
			}

			info = append(info, listData)
		} else {
			log.Printf("%s: %v", FormatVersionName, parsedData[FormatVersionName])

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
				log.Printf("yaml.Unmarshal(%s, %+v); Error: %s",
					rpmInfo, &listData, err.Error())
			}
			t, err := parseDate(parsedData["Build Date"])
			if err != nil {
				log.Printf("Build date: %+v\n", t)
				// listData.BuildDate = t
			}
			if "" != productVersion {
				allVersionsInfo := struct {
					VersionInfo []struct {
						Version          string `yaml:"product-version"`
						v2productVersion `yaml:",inline"`
					} `yaml:"compatibility-info"`
				}{}

				err := yaml.Unmarshal([]byte(rpmInfo), &allVersionsInfo)
				if err != nil {
					log.Printf("yaml.Unmarshal(%s, %+v); Error: %s",
						rpmInfo, &allVersionsInfo, err.Error())
				}

				// INFO: First check Version as-is,
				// 	if there is no match, then do pattern comparison.
				for _, vInfo := range allVersionsInfo.VersionInfo {
					if productVersion == vInfo.Version {
						listData.v2productVersion = vInfo.v2productVersion
						listData.matchedVersion = vInfo.Version
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

//  registerCommandList registers the list command that enables one to
// 	view the RPMs present in the software update repository.
func registerCommandList(progname string) {
	log.Printf("Entering repo::registerCommandList(%s)", progname)
	defer log.Println("Exiting repo::registerCommandList")

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
	output.RegisterCommandOptions(cmdOptions.listCmd,
		map[string]string{"output-format": "yaml"})
}
