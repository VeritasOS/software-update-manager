// Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package repo defines software repository functions like listing, removing
// 	packages from software repository.
package repo

import (
	"reflect"
	"testing"
	"time"
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

func Test_parseDate(t *testing.T) {
	wantT1, _ := time.Parse("Mon 02 Jan 2006 03:04:05 PM MST",
		"Wed 06 Jan 2021 04:48:21 PM PST")
	wantT2, _ := time.Parse(time.ANSIC, "Thu Feb  4 22:08:16 2021")
	type args struct {
		rawDate string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "Build date layout 1",
			args: args{
				rawDate: "Wed 06 Jan 2021 04:48:21 PM PST",
			},
			want: wantT1,
		},
		{
			name: "Build date layout 2",
			args: args{
				rawDate: "Thu Feb  4 22:08:16 2021",
			},
			want: wantT2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDate(tt.args.rawDate)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
