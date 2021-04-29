// Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package validate

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

var mockedExitStatus = 0
var mockedStdout string

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestExecCommandHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	es := strconv.Itoa(mockedExitStatus)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
		"STDOUT=" + mockedStdout,
		"EXIT_STATUS=" + es}
	return cmd
}

func TestExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stdout, os.Getenv("STDOUT"))
	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}

func Test_verifySign(t *testing.T) {
	type args struct {
		rpmFile    string
		errorMsg   string
		exitStatus int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid signature",
			args: args{
				rpmFile:    "asum.rpm",
				errorMsg:   "",
				exitStatus: 0,
			},
			wantErr: false,
		},
		{
			name: "Invalid signature",
			args: args{
				rpmFile: "asum.rpm",
				errorMsg: "Signature validation failed for asum.rpm. Only " +
					"install updates that have been downloaded from or " +
					"provided by Veritas. Error exit status 1",
				exitStatus: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockedExitStatus = tt.args.exitStatus
			mockedStdout = ""
			execCommand = fakeExecCommand
			defer func() { execCommand = exec.Command }()

			err := verifySign(tt.args.rpmFile)
			if (err != nil) && (tt.wantErr) && (err.Error() != tt.args.errorMsg) {
				t.Errorf("expected: <%v>, but got <%v>", tt.args.errorMsg, err.Error())
			}
		})
	}
}

func Test_isSigned(t *testing.T) {
	type args struct {
		rpmFile    string
		errorMsg   string
		exitStatus int
		stdout     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "RPM is signed",
			args: args{
				rpmFile:    "asum.rpm",
				errorMsg:   "",
				exitStatus: 0,
				stdout:     "RSA/SHA1, Wed 06 Feb 2019 06:50:08 PM PST, Key ID cf784714d9712e70",
			},
			wantErr: false,
		},
		{
			name: "RPM is signed",
			args: args{
				rpmFile:    "asum (1).rpm",
				errorMsg:   "",
				exitStatus: 0,
				stdout:     "RSA/SHA1, Wed 06 Feb 2019 06:50:08 PM PST, Key ID cf784714d9712e70",
			},
			wantErr: false,
		},
		{
			name: "RPM command throws error",
			args: args{
				rpmFile:    "asum.rpm",
				errorMsg:   "Failed to execute RPM command rpm -qip 'asum.rpm' | grep Signature | cut -d ':' -f 2-. Error exit status 1",
				exitStatus: 1,
				stdout:     "",
			},
			wantErr: true,
		},
		{
			name: "RPM file is not signed",
			args: args{
				rpmFile:    "asum.rpm",
				errorMsg:   "RPM file asum.rpm is not signed. Only install updates that have been downloaded from or provided by Veritas",
				exitStatus: 0,
				stdout:     "(none)",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockedExitStatus = tt.args.exitStatus
			mockedStdout = tt.args.stdout
			execCommand = fakeExecCommand
			defer func() { execCommand = exec.Command }()

			err := isSigned(tt.args.rpmFile)
			if (err != nil) && (tt.wantErr) && (err.Error() != tt.args.errorMsg) {
				t.Errorf("expected: <%v>, but got <%v>", tt.args.errorMsg, err.Error())
			}
		})
	}
}
