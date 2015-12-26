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
	"os"
	"strings"
	"unicode/utf8"
)

var chars = []byte{'-', '\\', '|', '/'}

type Spinner struct {
	index int
	len   int
}

func NewSpinner() *Spinner {
	return &Spinner{}
}

func (s *Spinner) Spin(text string) {
	bs := strings.Repeat("\b", s.len)
	sp := strings.Repeat(" ", s.len)
	fmt.Fprintf(os.Stderr, "%s%s%s%c %s", bs, sp, bs, chars[s.index], text)
	s.len = utf8.RuneCountInString(text) + 2
	s.index = (s.index + 1) % len(chars)
}

func (s *Spinner) Stop() {
	bs := strings.Repeat("\b", s.len)
	sp := strings.Repeat(" ", s.len)
	fmt.Fprintf(os.Stderr, "%s%s%s", bs, sp, bs)
	s.index = 0
	s.len = 0
}
