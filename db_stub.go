// +build nosqlite

package main

import "os"

func Get(string, os.FileInfo) (uint64, bool) { return 0, false }

func Upsert(string, os.FileInfo, uint64) {}

func Cleanup() {}
