package main

import (
	"reflect"
	"testing"
)

var testCases = []struct {
	in  string
	out []string
}{
	{"arg", []string{"arg"}},
	{"arg1 arg2", []string{"arg1", "arg2"}},
	{`"arg1 arg2"`, []string{"arg1 arg2"}},
	{"'arg1 arg2'", []string{"arg1 arg2"}},
	{`"arg1 arg2" arg3 arg4`, []string{"arg1 arg2", "arg3", "arg4"}},
	{`"arg1 arg2" arg3 arg4`, []string{"arg1 arg2", "arg3", "arg4"}},
	{`arg0 "arg1 arg2" arg3 arg4`, []string{"arg0", "arg1 arg2", "arg3", "arg4"}},
	{`arg0 "\"arg1\" 'arg2'" arg3 arg4`, []string{"arg0", `"arg1" 'arg2'`, "arg3", "arg4"}},
	{`arg1"arg2`, []string{`arg1"arg2`}},
	{"arg1    arg2", []string{"arg1", "arg2"}},
	{`"arg1    arg2"`, []string{`arg1    arg2`}},
	{`\"arg1\"`, []string{`"arg1"`}},
	{`\'arg1\'`, []string{`'arg1'`}},
	{"    arg1    ", []string{"arg1"}},
}

func TestParseArgs(t *testing.T) {
	for _, tc := range testCases {
		args := parseArgs(tc.in)
		if !reflect.DeepEqual(args, tc.out) {
			t.Errorf("parseArgs(%s): want %#v, got %#v", tc.in, tc.out, args)
		}
	}
}
