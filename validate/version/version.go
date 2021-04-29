// Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package version

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// V1VersionInfo would be decoded from the V1VersionInfo JSON array.
type V1VersionInfo struct {
	Version  string `json:"Version"`
	Reboot   string `json:"Reboot"`
	Estimate struct {
		Hours   string `json:"hours"`
		Minutes string `json:"minutes"`
		Seconds string `json:"seconds"`
	} `json:"Estimate"`
}

// Compare checks whether the specified version (with '*' patterns)
// 	matches the given product version.
func Compare(productVersion string, version string) bool {
	log.Printf("Entering version::Compare(%s, %s)", productVersion, version)
	defer log.Println("Exiting version::Compare")

	var num1, num2 string
	productVersionNums := strings.Split(productVersion, ".")
	versionNums := strings.Split(version, ".")
	maxLen := len(productVersionNums)

	if len(versionNums) > maxLen {
		maxLen = len(versionNums)
	}

	for i := 0; i < maxLen; i++ {
		num1, num2 = "0", "0"
		if i < len(productVersionNums) {
			num1 = productVersionNums[i]
		}
		if i < len(versionNums) {
			num2 = versionNums[i]
		}
		if num2 == "*" {
			return true
		}
		if num1 != num2 {
			return false
		}
	}
	return true
}

func validateJSONFormat(versionInfoString string) ([]V1VersionInfo, error) {
	log.Printf("Entering version::validateJSONFormat(%s)...",
		versionInfoString)
	defer log.Printf("Exiting version::validateJSONFormat...")

	versionInfoArray := make([]V1VersionInfo, 0)

	err := json.Unmarshal([]byte(versionInfoString), &versionInfoArray)
	if err != nil {
		log.Printf("json.Unmarshal(%s, %v); Error: %s",
			versionInfoString, &versionInfoArray, err.Error())
		err = fmt.Errorf("RPM V1VersionInfo is not in valid JSON format: %v", err)
		return versionInfoArray, err
	}
	return versionInfoArray, nil
}

func validateVersion(productVersion string, versionInfoArray []V1VersionInfo) (V1VersionInfo, error) {
	log.Printf("Entering version::validateVersion(%s, %s)...",
		productVersion, versionInfoArray)
	defer log.Println("Exiting version::validateVersion")

	versionSet := map[string]bool{}
	info := V1VersionInfo{}

	for _, versionInfo := range versionInfoArray {
		if versionSet[versionInfo.Version] {
			err := fmt.Errorf("Update version is not compatible for "+
				"the product version %v.", productVersion)
			log.Printf("Error in ValidateVersion: %s",
				err)
			return info, err
		}
		versionSet[versionInfo.Version] = true
	}

	for _, versionInfo := range versionInfoArray {
		if productVersion == versionInfo.Version {
			return versionInfo, nil
		}
		if Compare(productVersion, versionInfo.Version) {
			info = versionInfo
		}
	}

	if (info == V1VersionInfo{}) {
		err := fmt.Errorf("Update version is not "+
			"compatible for the product version %s.", productVersion)
		log.Printf("Error in ValidateVersion: %s",
			err)

		return V1VersionInfo{}, err
	}

	return info, nil
}

// GetCompatibileVersionInfo checks whether the product version is in the
// 	compatibility list, and returns version info for that product version.
func GetCompatibileVersionInfo(productVersion, versionInfoString string) (V1VersionInfo, error) {
	log.Printf("Entering version::GetCompatibileVersionInfo(%s, %s)...",
		productVersion, versionInfoString)
	defer log.Printf("Exiting version::GetCompatibileVersionInfo...")

	info := V1VersionInfo{}
	versionInfoArray, err := validateJSONFormat(versionInfoString)
	if err != nil {
		return info, err
	}

	info, err = validateVersion(productVersion, versionInfoArray)
	if err != nil {
		return info, err
	}

	return info, nil
}
