// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General
// Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	stdlog "log"
)

type quietVar int

func (q quietVar) String() string {
	return fmt.Sprint(int(q))
}

func (q quietVar) IsBoolFlag() bool {
	return true
}

func (q *quietVar) Set(val string) error {
	*q++
	return nil
}

func (q quietVar) Warn(v ...interface{}) {
	if q < 1 {
		stdlog.Print(v...)
	}
}

func (q quietVar) Warnf(format string, v ...interface{}) {
	if q < 1 {
		stdlog.Printf(format, v...)
	}
}

func (q quietVar) Error(v ...interface{}) {
	if q < 2 {
		stdlog.Print(v...)
	}
}

func (q quietVar) Errorf(format string, v ...interface{}) {
	if q < 2 {
		stdlog.Printf(format, v...)
	}
}

func (q quietVar) Fatal(v ...interface{}) {
	stdlog.Fatal(v...)
}

func (q quietVar) Fatalf(format string, v ...interface{}) {
	stdlog.Fatalf(format, v...)
}
