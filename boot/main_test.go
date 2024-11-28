/**
* Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9
 */

package boot

import (
	"fmt"
	"reflect"
	"testing"
)

var (
	releaseFile       string
	newVersion        string
	pathCurrent       map[string]string
	pathUpgrade       map[string]string
	kernelCurrent     string
	kernelUpgrade     string
	kernelPatch       string
	deviceCurrent     string
	deviceUpgrade     string
	kernelInfoCurrent KernelInfo
	kernelInfoUpgrade KernelInfo
	kernelInfoPatch   KernelInfo
)

func init() {
	releaseFile = "/etc/vxos-release"
	newVersion = "1.4.1"
	pathCurrent = map[string]string{"path": "./test_files/"}
	pathUpgrade = map[string]string{"path": "./test_files/system/upgrade/volume"}
	kernelCurrent = "3.10.0-1062.7.1.el7.x86_64"
	kernelUpgrade = "3.10.1-1062.7.1.el7.x86_64"
	kernelPatch = "3.10.0-1062.9.1.el7.x86_64"
	deviceCurrent = "/dev/mapper/system-thunder_cloud_1.4--999"
	deviceUpgrade = "/dev/mapper/system-thunder_cloud_2.0--999"
	kernelInfoCurrent = KernelInfo{Name: "current", Version: "1.4", Kernel: kernelCurrent, Device: deviceCurrent}
	kernelInfoPatch = KernelInfo{Name: "patch", Version: newVersion, Kernel: kernelPatch, Device: deviceCurrent}
	kernelInfoUpgrade = KernelInfo{Name: "upgrade", Version: "2.0", Kernel: kernelUpgrade, Device: deviceUpgrade}
}

func TestGetVersionInfo(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		version string
	}{
		{
			name:    "Use mocked release file",
			file:    "./test_files/" + releaseFile,
			version: "1.4",
		},
		{
			name:    "Use a non-existing file",
			file:    "./test_files/vxos-release",
			version: "",
		},
		{
			name:    "Use file without information of prodcut version",
			file:    "./test_files/incorrect-release",
			version: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			version := getVersionInfo(test.file)
			if version != test.version {
				t.Errorf("expected: [%v], but got [%v]", test.version, version)
			}
		})
	}
}

func TestGetKernelMap(t *testing.T) {
	type mock struct {
		device  string
		kernels string
	}

	tests := []struct {
		name        string
		mock        map[string]mock
		bootVolumes map[string]map[string]string
		newVersion  string
		kernelMap   []KernelInfo
	}{
		{
			name: "Kick Start",
			mock: map[string]mock{
				pathCurrent["path"]: mock{
					device:  deviceCurrent,
					kernels: kernelCurrent,
				},
			},
			bootVolumes: map[string]map[string]string{"current": pathCurrent},
			newVersion:  "",
			kernelMap:   []KernelInfo{kernelInfoCurrent},
		},
		{
			name: "Factory Reset",
			mock: map[string]mock{
				pathCurrent["path"]: mock{
					device:  deviceCurrent,
					kernels: kernelCurrent,
				},
				pathUpgrade["path"]: mock{
					device:  deviceUpgrade,
					kernels: kernelUpgrade,
				},
			},
			bootVolumes: map[string]map[string]string{"upgrade": pathUpgrade},
			newVersion:  "",
			kernelMap:   []KernelInfo{kernelInfoUpgrade},
		},
		{
			name: "Upgrade",
			mock: map[string]mock{
				pathCurrent["path"]: mock{
					device:  deviceCurrent,
					kernels: kernelCurrent,
				},
				pathUpgrade["path"]: mock{
					device:  deviceUpgrade,
					kernels: kernelUpgrade,
				},
			},
			bootVolumes: map[string]map[string]string{"current": pathCurrent, "upgrade": pathUpgrade},
			newVersion:  "",
			kernelMap:   []KernelInfo{kernelInfoCurrent, kernelInfoUpgrade},
		},
		{
			name: "Patch",
			mock: map[string]mock{
				pathCurrent["path"]: mock{
					device:  deviceCurrent,
					kernels: kernelCurrent + "\n" + kernelPatch,
				},
				pathUpgrade["path"]: mock{
					device:  deviceUpgrade,
					kernels: kernelUpgrade,
				},
			},
			bootVolumes: map[string]map[string]string{"current": pathCurrent},
			newVersion:  newVersion,
			kernelMap:   []KernelInfo{kernelInfoCurrent, kernelInfoPatch},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fExecCommand = func(operation string, command []string) string {
				return fmt.Sprintf("%v\n", kernelCurrent)
			}

			fGetVolumeInfo = func(volumePath string) (string, string) {
				return test.mock[volumePath].device, test.mock[volumePath].kernels
			}

			kernelMap := getKernelMap(releaseFile, test.bootVolumes, test.newVersion)
			if !reflect.DeepEqual(kernelMap, test.kernelMap) {
				t.Errorf("expected: [%v], but got [%v]", test.kernelMap, kernelMap)
			}
		})
	}
}
