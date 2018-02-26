// Copyright (c) 2018 Larry Hunter <larhun.it@gmail.com>. All rights reserved.
//
// Use of this source code is governed by a BSD 3-Clause license that can be
// found in the LICENSE file.

package golden

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// match represents a matching test. All error messages are accumulated and
// dumped when done. With one only error it passes whenever the found error
// matches the expected error (used for inner testing).
type match struct {
	*testing.T

	wantFail *string  // expected error (used for inner testing)
	messages []string // accumulated error messages
}

// newMatch returns a new matching test with the specified fail message.
func newMatch(t *testing.T, wantFail *string) *match {
	return &match{
		T:        t,
		wantFail: wantFail,
	}
}

// fail dumps the error list and fails the match.
func (m *match) fail() {
	m.Helper()
	for _, message := range m.messages {
		m.Error(message)
	}
	m.FailNow()
}

// done ends the match.
func (m *match) done() {
	m.Helper()

	switch {

	case m.wantFail == nil:
		if len(m.messages) != 0 {
			m.fail()
		}

	case len(m.messages) == 1:
		if m.match("WantFail", m.messages[0], *m.wantFail); len(m.messages) != 1 {
			m.fail() // expected fail error do not match current error
		}

	default:
		if m.match("WantFail", "", *m.wantFail); len(m.messages) != 0 {
			m.fail() // expected fail error not found
		}
	}

	return
}

// equal tests if the got value is equal to the want value. If not, accumulates
// an error with the specified name.
func (m *match) equal(name string, got, want int) {
	if got != want {
		message := name + " match error:\ngot: " + strconv.Itoa(got) + ", want: " + strconv.Itoa(want)
		m.messages = append(m.messages, message)
	}
}

// match tests if the got string matches the want smart string. If not,
// accumulates an error with the specified name.
func (m *match) match(name, got, want string) {
	n := len(want)
	ext := filepath.Ext(want)
	ok := false

	switch {

	case want == "golden"+ext:
		name += " golden" + ext
		if want, ok = m.getGolden(name, ext, got); !ok {
			return // file error
		}
		ok = got == want

	case n > 2 && want[0] == '^' && want[n-1] == '$':
		if re, err := regexp.Compile(want); err == nil {
			name += " pattern"
			ok = re.MatchString(got)
		} else {
			ok = got == want
		}

	case n > 6 && want[0:3] == "..." && want[n-3:n] == "...":
		name += " substring"
		ok = strings.Contains(got, want[3:len(want)-3])

	case n > 3 && want[n-3:n] == "...":
		name += " prefix"
		ok = strings.HasPrefix(got, want[:len(want)-3])

	case n > 3 && want[0:3] == "...":
		name += " suffix"
		ok = strings.HasSuffix(got, want[3:])

	case n > 1 && want[0] == '=':
		name += " escaped"
		ok = got == want[1:]

	default:
		ok = got == want
	}

	if !ok {
		m.messages = append(m.messages, format(name, got, want))
	}

	return
}

// getGolden returns the content of the named gold master file with the
// specified extension and reports if succeeded. If the update flag is true,
// writes the got string before reading.
func (m *match) getGolden(name, ext, got string) (string, bool) {
	file := filepath.Join(goldenDir, strings.Replace(m.Name(), "/", "-", -1)+ext)

	if *update {
		if err := os.MkdirAll(goldenDir, 0700); err != nil {
			m.messages = append(m.messages, name+" folder error:\n"+err.Error())
			return "", false
		}

		if err := ioutil.WriteFile(file, []byte(got), 0600); err != nil {
			m.messages = append(m.messages, name+" update error:\n"+err.Error())
			return "", false
		}
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		m.messages = append(m.messages, name+" read error:\n"+err.Error())
		return "", false
	}

	return string(data), true
}

// getFile returns the content and extension values of the named file and
// reports if succeeded.
func (m *match) getFile(name string) (got, ext string, ok bool) {
	ext = filepath.Ext(name)

	data, err := ioutil.ReadFile(name)
	if err != nil {
		m.messages = append(m.messages, "File read error:\n"+err.Error())
		return "", "", false
	}

	got = string(data)
	return got, ext, true
}

// removeFile removes the named file and reports if succeeded.
func (m *match) removeFile(name string) bool {
	stat, err := os.Stat(name)
	if os.IsNotExist(err) {
		return true
	}

	if err != nil {
		m.messages = append(m.messages, "File access error:\n"+err.Error())
		return false
	}

	if !stat.Mode().IsRegular() {
		m.messages = append(m.messages, "File mode error:\nexpected regular file")
		return false
	}

	if err := os.Remove(name); err != nil {
		m.messages = append(m.messages, "File remove error:\n"+err.Error())
		return false
	}

	return true
}

// run runs the function and returns the panic message, if any.
func (m *match) run(f func()) (message string) {
	defer func() {
		if r := recover(); r != nil {
			message = fmt.Sprint(r)
		}
	}()
	f()
	return
}
