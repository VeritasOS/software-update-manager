// Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package repo defines software repository functions like listing, removing
// 	packages from software repository.
package rpm

import (
	"reflect"
	"testing"
)

func Test_ParseMetaData(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "V1 RPM output",
			args: args{
				data: `Name        : platformx-upgrade
Version     : 3.3.13
Release     : 20200325132333
Architecture: x86_64
Install Date: (not installed)
Group       : Unspecified
Size        : 4150750911
License     : Copyright (c) 2019 Veritas Technologies LLC. All rights reserved.
Signature   : RSA/SHA256, Wed 25 Mar 2020 03:05:55 PM PDT, Key ID cf784714d9712e70
Source RPM  : platformx-upgrade-3.3.13-20200325132333.src.rpm
Build Date  : Wed 25 Mar 2020 02:42:20 PM PDT
Build Host  : k8s-prod-rhel72-10g-7tdz2
Relocations : /system/upgrade/repository/platformx_core 
Packager    : Veritas Technologies LLC
Vendor      : Veritas Technologies LLC
URL         : https://www.veritas.com/support/en_US.html
Summary     : Provides platformx-upgrade.
Description :
Platform Upgrade Package. Will reboot system.
Type:Upgrade
VersionInfo:[{"Version":"0.0.9","Reboot":"Yes","Estimate":{"hours":"0","minutes":"20","seconds":"15"}}]
`,
			},
			want: map[string]string{
				"Architecture": "x86_64",
				"Install Date": "(not installed)",
				"Description":  "Platform Upgrade Package. Will reboot system.",
				"Group":        "Unspecified",
				"License":      "Copyright (c) 2019 Veritas Technologies LLC. All rights reserved.",
				"Name":         "platformx-upgrade",
				"Packager":     "Veritas Technologies LLC",
				"Release":      "20200325132333",
				"Relocations":  "/system/upgrade/repository/platformx_core",
				"Signature":    "RSA/SHA256, Wed 25 Mar 2020 03:05:55 PM PDT, Key ID cf784714d9712e70",
				"Source RPM":   "platformx-upgrade-3.3.13-20200325132333.src.rpm",
				"Build Date":   "Wed 25 Mar 2020 02:42:20 PM PDT",
				"Build Host":   "k8s-prod-rhel72-10g-7tdz2",
				"Size":         "4150750911",
				"Summary":      "Provides platformx-upgrade.",
				"Type":         "Upgrade",
				"URL":          "https://www.veritas.com/support/en_US.html",
				"Vendor":       "Veritas Technologies LLC",
				"Version":      "3.3.13",
				"VersionInfo":  `[{"Version":"0.0.9","Reboot":"Yes","Estimate":{"hours":"0","minutes":"20","seconds":"15"}}]`,
			},
		},
		{
			name: "V2 RPM output",
			args: args{
				data: `Name        : VRTSasum-update
Version     : 2.0.1
Release     : 20200723001743
Architecture: x86_64
Install Date: (not installed)
Group       : Unspecified
Size        : 6098439
License     : Copyright (c) 2020 Veritas Technologies LLC. All rights reserved.
Signature   : RSA/SHA256, Mon 06 Jul 2020 02:41:26 PM PDT, Key ID cf784714d9712e70
Source RPM  : VRTSasum-update-2.0.1-20200723001743.src.rpm
Build Date  : Wed 22 Jul 2020 05:17:50 PM PDT
Build Host  : builder
Relocations : (not relocatable)
Packager    : Veritas AS DevOps <DL-VTAS-AS-Team-Peregrine@veritas.com>
URL         : https://www.veritas.com/support/en_US.html
Summary     : Sample Update RPM created using ASUM SDK
Description :
RPM Format Version : 2
RPM Info    : {"Description":["Sample multi-line description of the update RPM.","Client scripts/GUI can display each item in this list in a separate para with appropriate linebreaks."],"Type":"Update","VersionInfo":[{"Version":"*","install":{"confirmation-message":["Sample multi-line confirmation message","","Display info messages and instructions like restarting node, ","and confirming that users have stopped instances."],"requires-restart":true,"supports-rollback":false,"estimated-minutes":35},"rollback":{"confirmation-message":["Sample multi-line confirmation message","","Display info messages and instructions like restarting node, ","and confirming that users have stopped instances."],"requires-restart":true,"estimated-minutes":20},"commit":{"confirmation-message":["Sample multi-line confirmation message","","Display info messages and instructions like once update is committed, you cannot roll back.","Also, committing node restarts some services."],"estimated-minutes":5}},{"Version":"3.*","install":{"confirmation-message":["Sample multi-line confirmation message","","Display warning messages and instructions like stopping instances."],"requires-restart":false,"supports-rollback":false,"estimated-minutes":25},"rollback":{"confirmation-message":["Sample multi-line confirmation message","","Display info messages and instructions like roll back requires restarting node, ","(even though install didn't require restart), as snapshot needs to be reverted."],"requires-restart":true,"estimated-minutes":40},"commit":{"confirmation-message":["Sample multi-line confirmation message","","Display info messages and instructions like once update is committed, you cannot roll back."],"estimated-minutes":5}}]}
`,
			},
			want: map[string]string{
				"RPM Format Version": "2",
				"Architecture":            "x86_64",
				"Build Date":              "Wed 22 Jul 2020 05:17:50 PM PDT",
				"Build Host":              "builder",
				"Description":             "",
				"Group":                   "Unspecified",
				"Install Date":            "(not installed)",
				"License":                 "Copyright (c) 2020 Veritas Technologies LLC. All rights reserved.",
				"Name":                    "VRTSasum-update",
				"Packager":                "Veritas AS DevOps <DL-VTAS-AS-Team-Peregrine@veritas.com>",
				"RPM Info":                `{"Description":["Sample multi-line description of the update RPM.","Client scripts/GUI can display each item in this list in a separate para with appropriate linebreaks."],"Type":"Update","VersionInfo":[{"Version":"*","install":{"confirmation-message":["Sample multi-line confirmation message","","Display info messages and instructions like restarting node, ","and confirming that users have stopped instances."],"requires-restart":true,"supports-rollback":false,"estimated-minutes":35},"rollback":{"confirmation-message":["Sample multi-line confirmation message","","Display info messages and instructions like restarting node, ","and confirming that users have stopped instances."],"requires-restart":true,"estimated-minutes":20},"commit":{"confirmation-message":["Sample multi-line confirmation message","","Display info messages and instructions like once update is committed, you cannot roll back.","Also, committing node restarts some services."],"estimated-minutes":5}},{"Version":"3.*","install":{"confirmation-message":["Sample multi-line confirmation message","","Display warning messages and instructions like stopping instances."],"requires-restart":false,"supports-rollback":false,"estimated-minutes":25},"rollback":{"confirmation-message":["Sample multi-line confirmation message","","Display info messages and instructions like roll back requires restarting node, ","(even though install didn't require restart), as snapshot needs to be reverted."],"requires-restart":true,"estimated-minutes":40},"commit":{"confirmation-message":["Sample multi-line confirmation message","","Display info messages and instructions like once update is committed, you cannot roll back."],"estimated-minutes":5}}]}`,
				"Release":                 "20200723001743",
				"Relocations":             "(not relocatable)",
				"Signature":               "RSA/SHA256, Mon 06 Jul 2020 02:41:26 PM PDT, Key ID cf784714d9712e70",
				"Size":                    "6098439",
				"Source RPM":              "VRTSasum-update-2.0.1-20200723001743.src.rpm",
				"Summary":                 "Sample Update RPM created using ASUM SDK",
				"URL":                     "https://www.veritas.com/support/en_US.html",
				"Version":                 "2.0.1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseMetaData(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMetaData() = %v, want %v", got, tt.want)
			}
		})
	}
}
