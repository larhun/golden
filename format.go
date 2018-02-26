// Copyright (c) 2018 Larry Hunter <larhun.it@gmail.com>. All rights reserved.
//
// Use of this source code is governed by a BSD 3-Clause license that can be
// found in the LICENSE file.

package golden

import (
	"bytes"
	"strings"
)

// isMultiline reports whether the string is multiline. An ending newline is
// ignored.
func isMultiline(s string) bool {
	for i := 0; i+1 < len(s); i++ {
		if s[i] == '\n' {
			return true
		}
	}
	return false
}

// format returns a formatted error message that reports the test name and the
// got and want values.
func format(name, got, want string) string {
	m := new(message)

	m.WriteString(name)
	m.WriteString(" match error:")
	switch {

	case strings.HasPrefix(name, "WantFail"):
		m.WriteString(" expected: ")
		m.WriteString(want)

	case got == "":
		m.WriteString("\ngot an empty string, want:")
		if isMultiline(want) {
			m.WriteIndent(want)
		} else {
			m.WriteString(" ")
			m.WriteLine(want)
		}

	case want == "":
		m.WriteString("\nwant an empty string, got:")
		if isMultiline(got) {
			m.WriteIndent(got)
		} else {
			m.WriteString(" ")
			m.WriteLine(got)
		}

	case isMultiline(got) || isMultiline(want):
		m.WriteString("\ngot:")
		m.WriteIndent(got)
		m.WriteString("\nwant:")
		m.WriteIndent(want)

	default:
		m.WriteString("\ngot: ")
		m.WriteLine(got)
		m.WriteString("\nwant: ")
		m.WriteLine(want)
	}

	return m.String()
}

// message represents a message writer.
type message struct {
	bytes.Buffer
}

// WriteLine writes the string as one line. An ending newline is ignored.
func (m *message) WriteLine(s string) {
	if n := len(s); n > 0 && s[n-1] == '\n' {
		m.WriteString(s[:n-1])
	} else {
		m.WriteString(s)
	}
}

// WriteIndent writes the string indenting all lines by four spaces.
func (m *message) WriteIndent(s string) {
	i := 0
	for j := 0; j < len(s); j++ {
		if s[j] == '\n' {
			m.WriteString("\n    ")
			m.WriteString(s[i:j])
			i = j + 1
		}
	}
	m.WriteString("\n    ")
	m.WriteString(s[i:])
}
