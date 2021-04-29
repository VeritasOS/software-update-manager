// Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package repo defines software repository functions like listing, removing
// 	packages from software repository.
package repo

import (
	"os"
	logutil "plugin-manager/utils/log"
	"strings"
	"testing"
)

func TestRemove(t *testing.T) {
	type args struct {
		swName string
		swType string
		swRepo string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Repo not specified",
			args:    args{},
			wantErr: true,
			errMsg:  "Failed to determine software repository.",
		},
		{
			name: "Only Repo specified",
			args: args{
				swRepo: "/tmp/software/repository/",
			},
			wantErr: false,
		},
		{
			name: "Repo & Type specified",
			args: args{
				swRepo: "/tmp/software/repository/",
				swType: "update",
			},
			wantErr: false,
		},
		{
			name: "Repo & Name specified, but not Type",
			args: args{
				swRepo: "/tmp/software/repository/",
				swName: "ab.rpm",
			},
			wantErr: true,
			errMsg:  "Software type must be specified when software name is specified.",
		},
		{
			name: "Repo, Type & Name specified",
			args: args{
				swRepo: "/tmp/software/repository/",
				swType: "update",
				swName: "ab.rpm",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			touchRpm := false
			if tt.args.swRepo != "" {
				path := tt.args.swRepo
				if tt.args.swType != "" {
					path += string(os.PathSeparator) + tt.args.swType
					if tt.args.swName != "" {
						touchRpm = true
					}
				}
				err = os.MkdirAll(path, 0766)
				if err != nil {
					t.Errorf("Failed to create repository %s. Error: %s.",
						path, err.Error())
				}

				if touchRpm {
					path += string(os.PathSeparator) + tt.args.swName
					fi, err := os.Create(path)
					if err != nil {
						t.Errorf("Failed to create new file %s. Error: %s.", path, err.Error())
					}
					logutil.PrintNLog("New file details: %+v\n", fi.Name())
					fi.Close()
				}
			}
			if err = Remove(tt.args.swName, tt.args.swType, tt.args.swRepo); (err != nil) != tt.wantErr {
				t.Errorf("Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Remove() error = %v, wantErr %v", err.Error(), tt.errMsg)
			}
		})
	}
}
