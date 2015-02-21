package main

import (
	"strings"
)

// StringSlice is a custom flag that implements flag.Value interface.
// It can be specified multiple times on the command line and collects
// all the values in a slice.
type StringSlice []string

func (a *StringSlice) Set(val string) error {
	*a = append(*a, val)
	return nil
}

func (a *StringSlice) String() string {
	return strings.Join(*a, " ")
}
