// +build nosqlite

package main

import "time"

func Get(string, time.Time) (uint64, bool) { return 0, false }

func Upsert(string, time.Time, uint64) {}

func Cleanup() {}
