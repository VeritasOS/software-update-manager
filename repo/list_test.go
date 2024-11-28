// Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package repo defines software repository functions like listing, removing
// packages from software repository.
package repo

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	logger "github.com/VeritasOS/plugin-manager/utils/log"

	"gopkg.in/yaml.v3"
)

func TestList(t *testing.T) {
	type args struct {
		params map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]*interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := List(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func readFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	return ioutil.ReadAll(file)
}

// TODO: Fix this test
func _TestListRPMFilesInfo(t *testing.T) {
	logger.InitFileLogger("test.log", "INFO")

	type args struct {
		files            []string
		productVersion   string
		includeFields    string
		metadataFilePath string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "should return checksum alongwith default attributes",
			args: args{
				files:            []string{"VRTSflex-update-3.2.rpm"},
				productVersion:   "2.1",
				includeFields:    "checksum",
				metadataFilePath: "./test_files/valid-rpm-metadata.txt",
			},
			want: "./test_files/include-checksum-response.yaml",
		},
		{
			name: "Should return default fields if --include-fields not provided",
			args: args{
				files:            []string{"VRTSflex-update-3.2.rpm"},
				productVersion:   "2.1",
				includeFields:    "",
				metadataFilePath: "./test_files/valid-rpm-metadata.txt",
			},
			want: "./test_files/default-response.yaml",
		},
		{
			name: "Should return default fields if invalid --include-fields provided",
			args: args{
				files:            []string{"VRTSflex-update-3.2.rpm"},
				productVersion:   "2.1",
				includeFields:    "asas",
				metadataFilePath: "./test_files/valid-rpm-metadata.txt",
			},
			want: "./test_files/default-response.yaml",
		},
	}

	originalGetChecksum := fGetChecksum
	originalfGetRPMPackageInfo := fGetRPMPackageInfo
	defer func() {
		fGetChecksum = originalGetChecksum
		fOpen = os.Open
		fGetRPMPackageInfo = originalfGetRPMPackageInfo
	}()
	fGetChecksum = func(src io.Reader) string {
		return "random_checksum"
	}
	fOpen = func(name string) (*os.File, error) {
		return nil, nil
	}
	for _, tt := range tests {
		fGetRPMPackageInfo = func(rpmPath string) ([]byte, error) {
			return readFile(tt.args.metadataFilePath)
		}
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListRPMFilesInfo(tt.args.files, tt.args.productVersion, tt.args.includeFields)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			jsonOut, err := yaml.Marshal(got)
			if err != nil {
				panic(err)
			}
			expectedByte, _ := readFile(tt.want)

			ex := []byte{}
			g := []byte{}
			counter := -1

			for i, b := range expectedByte {
				if string(jsonOut[i]) != string(b) {
					ex = append(ex, jsonOut[i])
					g = append(g, b)
					counter = 0
				} else {
					if counter == 0 && len(g) != 0 {
						t.Errorf("Got =  %v, expected  %v", string(g), string(ex))
					}
					ex = ex[:0]
					g = g[:0]
					counter = -1
				}
			}
		})
	}
}
