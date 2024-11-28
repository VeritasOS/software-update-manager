/**
* Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9
 */

package boot

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// Variables for retrieving the version information from release file
const (
	VxosReleaseFile = "/etc/vxos-release"
	VersionPattern  = "product-version = "
)

// KernelInfo stores the information for kernels
type KernelInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Kernel  string `json:"kernel"`
	Device  string `json:"device"`
}

var (
	fExecCommand   = execCommand
	fGetVolumeInfo = getVolumeInfo
)

// Execution handler for running a command on os
func execCommand(operation string, command []string) string {
	out, err := exec.Command(operation, command...).Output()
	if err != nil {
		message := fmt.Sprintf("Failed to execute command '%s %s'. Error: %s.", operation, command, err)
		fmt.Fprintf(os.Stderr, "%v\n", message)
	}
	return string(out)
}

func execCommandBash(cmdStr string) (string, error) {
	cmd := exec.Command("bash", "-c", cmdStr)
	out, err := cmd.Output()
	outStr := string(out)
	if err != nil {
		message := fmt.Sprintf("Failed to run command [%s].", cmdStr)
		err = errors.New(message)
	}
	return outStr, err
}

func getVersionInfo(file string) string {
	content, _ := ioutil.ReadFile(file)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		if strings.Contains(line, VersionPattern) {
			return strings.Split(line, VersionPattern)[1]
		}
	}
	return ""
}

func getVolumeInfo(volumePath string) (string, string) {
	rpmCmd := "rpm -q kernel --root=" + volumePath + " --qf '%{VERSION}-%{RELEASE}.%{ARCH}\n'"
	device := strings.TrimSpace(execCommand("findmnt", []string{"-n", "-o", "SOURCE", volumePath}))
	kernels, err := execCommandBash(rpmCmd)
	if err != nil {
		return device, ""
	}
	return device, kernels
}

func getKernelMap(releaseFile string, bootVolumes map[string]map[string]string, newVersion string) []KernelInfo {
	kernelMap := make([]KernelInfo, 0)
	runningKernel := strings.TrimSpace(fExecCommand("uname", []string{"-r"}))

	volumeNames := make([]string, 0, len(bootVolumes))
	for volumeName := range bootVolumes {
		volumeNames = append(volumeNames, volumeName)
	}
	sort.Strings(volumeNames)

	for _, volumeName := range volumeNames {
		for _, volumePath := range bootVolumes[volumeName] {
			version := getVersionInfo(volumePath + releaseFile)
			device, kernels := fGetVolumeInfo(volumePath)

			for _, kernel := range strings.Fields(kernels) {
				kernelInfo := KernelInfo{
					Name:    volumeName,
					Version: version,
					Kernel:  kernel,
					Device:  device,
				}
				if volumeName == "current" && kernel != runningKernel {
					kernelInfo.Name = "patch"
					if len(newVersion) != 0 {
						kernelInfo.Version = newVersion
					}
				}
				kernelMap = append(kernelMap, kernelInfo)
			}
		}
	}
	return kernelMap
}

// Exec executes boot management
func Exec(args []string) error {
	os.Args = args

	prepareFlag := flag.Bool("prepare", false, "command for preparing boot info: boot -prepare -release_file=<filename> -volumes=<json> -version=<version>")
	releaseFile := flag.String("release_file", "", "path of release file")
	bootVolumes := flag.String("volumes", "", "bootable volumes")
	newVersion := flag.String("version", "", "version")

	flag.Parse()

	if *prepareFlag {
		if *bootVolumes == "" {
			flag.PrintDefaults()
			return fmt.Errorf("Bootable volumes need to be specified")
		}
		if *releaseFile == "" {
			*releaseFile = VxosReleaseFile
		}
		bootVolumesMap := make(map[string]map[string]string)
		json.Unmarshal([]byte(*bootVolumes), &bootVolumesMap)
		kernelMap := getKernelMap(*releaseFile, bootVolumesMap, *newVersion)
		kernelMapOutput, _ := json.Marshal(kernelMap)
		fmt.Println(string(kernelMapOutput))
		return nil
	}
	flag.PrintDefaults()
	return fmt.Errorf("Invalid inputs of command")
}
