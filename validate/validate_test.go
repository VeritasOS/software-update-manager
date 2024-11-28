// Copyright (c) 2022 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package validate

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	logger "github.com/VeritasOS/plugin-manager/utils/log"
)

type cmdRep struct {
	cmdOut   string
	exitCode int
}

var mockedCmdExec map[string]cmdRep

// Based on different command, return different result
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestExecCommandHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	fullCmd := command + " " + strings.Join(args, " ")
	es := strconv.Itoa(mockedCmdExec[fullCmd].exitCode)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
		"STDOUT=" + mockedCmdExec[fullCmd].cmdOut,
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
		rpmFile string            // function input
		cmdExec map[string]cmdRep // cmd input and response
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		errMsg  string
	}{
		{
			name: "Signature ealier than Gorden GA date, skip digest check", // # 1670198400
			args: args{
				rpmFile: "asum.rpm",
				cmdExec: map[string]cmdRep{
					"/bin/sh -c /bin/rpm -qp --queryformat '%{BUILDTIME}' 'asum.rpm'": {
						cmdOut:   "1670198000",
						exitCode: 0,
					},
					"/bin/rpm -Kv --nodigest asum.rpm": {
						cmdOut:   "",
						exitCode: 0,
					},
					"/bin/rpm -Kv asum.rpm": {
						cmdOut:   "",
						exitCode: 1,
					},
				},
			},
			errMsg:  "",
			wantErr: false,
		},
		{
			name: "Signature later than Gorden GA date, digest check succeeds",
			args: args{
				rpmFile: "asum.rpm",
				cmdExec: map[string]cmdRep{
					"/bin/sh -c /bin/rpm -qp --queryformat '%{BUILDTIME}' 'asum.rpm'": {
						cmdOut:   "1670200000",
						exitCode: 0,
					},
					"/bin/rpm -Kv --nodigest asum.rpm": {
						cmdOut:   "",
						exitCode: 1,
					},
					"/bin/rpm -Kv asum.rpm": {
						cmdOut:   "",
						exitCode: 0,
					},
				},
			},
			errMsg:  "",
			wantErr: false,
		},
		{
			name: "Signature later than Gorden GA date, digest check fails",
			args: args{
				rpmFile: "asum.rpm",
				cmdExec: map[string]cmdRep{
					"/bin/sh -c /bin/rpm -qp --queryformat '%{BUILDTIME}' 'asum.rpm'": {
						cmdOut:   "1670200000",
						exitCode: 0,
					},
					"/bin/rpm -Kv --nodigest asum.rpm": {
						cmdOut:   "",
						exitCode: 1,
					},
					"/bin/rpm -Kv asum.rpm": {
						cmdOut:   "",
						exitCode: 1,
					},
				},
			},
			errMsg:  "Signature validation failed for asum.rpm. Only install updates that have been downloaded from or provided by Veritas. Error exit status 1",
			wantErr: true,
		},
	}
	// Set log file name to "test", so that cleaning becomes easier.
	logger.InitFileLogger("test.log", "INFO")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockedCmdExec = tt.args.cmdExec
			execCommand = fakeExecCommand
			defer func() { execCommand = exec.Command }()

			err := verifySign(tt.args.rpmFile)
			if (err != nil) && (tt.wantErr) && (err.Error() != tt.errMsg) {
				t.Errorf("expected: <%v>, but got <%v>", tt.errMsg, err.Error())
			}
		})
	}
}

func Test_isSigned(t *testing.T) {
	type args struct {
		rpmFile string            // function input
		cmdExec map[string]cmdRep // cmd input and response
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		errMsg  string
	}{
		{
			name: "RPM is signed",
			args: args{
				rpmFile: "asum.rpm",
				cmdExec: map[string]cmdRep{
					"sh -c rpm -qip 'asum.rpm' | grep Signature | cut -d ':' -f 2-": {
						cmdOut:   "RSA/SHA1, Wed 06 Feb 2019 06:50:08 PM PST, Key ID cf784714d9712e70",
						exitCode: 0,
					},
				},
			},
			errMsg:  "",
			wantErr: false,
		},
		{
			name: "RPM command throws error",
			args: args{
				rpmFile: "asum.rpm",
				cmdExec: map[string]cmdRep{
					"sh -c rpm -qip 'asum.rpm' | grep Signature | cut -d ':' -f 2-": {
						cmdOut:   "",
						exitCode: 1,
					},
				},
			},
			errMsg:  "Failed to execute RPM command rpm -qip 'asum.rpm' | grep Signature | cut -d ':' -f 2-. Error exit status 1",
			wantErr: true,
		},
		{
			name: "RPM file is not signed",
			args: args{
				rpmFile: "asum.rpm",
				cmdExec: map[string]cmdRep{
					"sh -c rpm -qip 'asum.rpm' | grep Signature | cut -d ':' -f 2-": {
						cmdOut:   "(none)",
						exitCode: 0,
					},
				},
			},
			errMsg:  "RPM file asum.rpm is not signed. Only install updates that have been downloaded from or provided by Veritas",
			wantErr: true,
		},
	}
	// Set log file name to "test", so that cleaning becomes easier.
	logger.InitFileLogger("test.log", "INFO")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockedCmdExec = tt.args.cmdExec
			execCommand = fakeExecCommand
			defer func() { execCommand = exec.Command }()

			err := isSigned(tt.args.rpmFile)
			if (err != nil) && (tt.wantErr) && (err.Error() != tt.errMsg) {
				t.Errorf("expected: <%v>, but got <%v>", tt.errMsg, err.Error())
			}
		})
	}
}
