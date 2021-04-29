// Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package version

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCompare(t *testing.T) {
	type args struct {
		productVersion string
		version        string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "First digit - exact match",
			args: args{productVersion: "2", version: "2"},
			want: true,
		},
		{
			name: "First digit - pattern match",
			args: args{productVersion: "2", version: "*"},
			want: true,
		},
		{
			name: "First digit - no match",
			args: args{productVersion: "2", version: "3"},
			want: false,
		},
		{
			name: "Two digit - exact match",
			args: args{productVersion: "2.0", version: "2.0"},
			want: true,
		},
		{
			name: "Two digit - pattern match",
			args: args{productVersion: "2.0", version: "2.*"},
			want: true,
		},
		{
			name: "Two digit - no match",
			args: args{productVersion: "2.0", version: "2.2"},
			want: false,
		},
		{
			name: "Three digits - exact match",
			args: args{productVersion: "1.2.3", version: "1.2.3"},
			want: true,
		},
		{
			name: "Two and Four digits - pattern match",
			args: args{productVersion: "2.0", version: "2.0.0.*"},
			want: true,
		},
		{
			name: "Fourth digit pattern match",
			args: args{productVersion: "2.0.0.9", version: "2.0.0.*"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Compare(tt.args.productVersion, tt.args.version); got != tt.want {
				t.Errorf("Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateVersion(t *testing.T) {
	tests := []struct {
		name           string
		versionInfo    string
		version        string
		matchedVersion string
		wantErr        bool
	}{
		{
			name:           "Exact Match (Version:1.0.0)",
			versionInfo:    `[{"Version":"1.0.0"},{"Version":"1.0.1.1"},{"Version":"1.0.1.*"}]`,
			version:        "1.0.0",
			matchedVersion: "1.0.0",
			wantErr:        false,
		},
		{
			name:           "Exact Match (Version:1.0.1.1)",
			versionInfo:    `[{"Version":"1.0.0"},{"Version":"1.0.1.1"},{"Version":"1.0.1.*"}]`,
			version:        "1.0.1.1",
			matchedVersion: "1.0.1.1",
			wantErr:        false,
		},
		{
			name:           "Pattern Match (Version:1.0.1)",
			versionInfo:    `[{"Version":"1.0.0"},{"Version":"1.0.1.1"},{"Version":"1.0.1.*"}]`,
			version:        "1.0.1",
			matchedVersion: "1.0.1.*",
			wantErr:        false,
		},
		{
			name:           "Pattern Match (Version:1.0.1.)",
			versionInfo:    `[{"Version":"1.0.0"},{"Version":"1.0.1.1"},{"Version":"1.0.1.*"}]`,
			version:        "1.0.1.",
			matchedVersion: "1.0.1.*",
			wantErr:        false,
		},
		{
			name:           "Pattern Match (Version:1.0.1.2)",
			versionInfo:    `[{"Version":"1.0.0"},{"Version":"1.0.1.1"},{"Version":"1.0.1.*"}]`,
			version:        "1.0.1.2",
			matchedVersion: "1.0.1.*",
			wantErr:        false,
		},
		{
			name:           "Pattern Match (Version:1.0.1.1a)",
			versionInfo:    `[{"Version":"1.0.0"},{"Version":"1.0.1.1"},{"Version":"1.0.1.*"}]`,
			version:        "1.0.1.1a",
			matchedVersion: "1.0.1.*",
			wantErr:        false,
		},
		{
			name:           "Mismatch (Version:1.0.10)",
			versionInfo:    `[{"Version":"1.0.0"},{"Version":"1.0.1.1"},{"Version":"1.0.1.*"}]`,
			version:        "1.0.10",
			matchedVersion: "",
			wantErr:        true,
		},
		{
			name:           "Duplicate Info",
			versionInfo:    `[{"Version":"1.0.0"},{"Version":"1.0.0"},{"Version":"1.0.1.*"}]`,
			version:        "1.0.10",
			matchedVersion: "",
			wantErr:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			versionInfo := make([]V1VersionInfo, 0)
			json.Unmarshal([]byte(test.versionInfo), &versionInfo)
			output, err := validateVersion(test.version, versionInfo)
			if (err != nil) != test.wantErr {
				t.Log(err.Error())
				t.Errorf("expected: <%v>, but got <%v>", test.wantErr, err)
				return
			}

			if reflect.DeepEqual(output.Version, test.matchedVersion) == false {
				t.Errorf("expected: <%s>, but got <%s>",
					test.matchedVersion, output.Version)
			}
		})
	}
}
