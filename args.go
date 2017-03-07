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

import "bytes"

func parseArgs(s string) (args []string) {
	quoted := false
	canStartQuoted := true
	var quote byte
	var buf bytes.Buffer

	for i := 0; i < len(s); i++ {
		b := s[i]
		switch b {
		case '"', '\'':
			if quoted {
				if b == quote {
					quoted = false
				} else {
					buf.WriteByte(b)
				}
			} else if canStartQuoted {
				quoted = true
				quote = b
			} else {
				buf.WriteByte(b)
			}
		case ' ':
			if quoted {
				buf.WriteByte(' ')
			} else {
				if buf.Len() > 0 {
					args = append(args, buf.String())
					buf.Reset()
				}
				for i+1 < len(s) && s[i+1] == ' ' {
					i++
				}
				canStartQuoted = true
			}
		case '\\':
			if i+1 < len(s) && (s[i+1] == '"' || s[i+1] == '\'') {
				b = s[i+1]
				i++
			}

			fallthrough
		default:
			buf.WriteByte(b)
			canStartQuoted = false
		}
	}

	if buf.Len() > 0 {
		args = append(args, buf.String())
	}

	return
}
