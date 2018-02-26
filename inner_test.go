// Copyright (c) 2018 Larry Hunter <larhun.it@gmail.com>.
// All rights reserved.
//
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package golden

// Functions required to test the error printed by a test that should fail.

// FailCase is a Case with an exported WantFail field.
type FailCase struct {
	Name         string
	Args         []string
	WantFile     string
	WantStdout   string
	WantStderr   string
	WantPanic    string
	WantErr      string
	WantFail     *string // exported field
	WantExitCode int
}

// ToCase return a list of test cases.
func ToCase(failCases []FailCase) []Case {
	testCases := make([]Case, len(failCases))
	for i, fc := range failCases {
		testCases[i] = Case{
			Name:         fc.Name,
			Args:         fc.Args,
			WantFile:     fc.WantFile,
			WantStdout:   fc.WantStdout,
			WantStderr:   fc.WantStderr,
			WantPanic:    fc.WantPanic,
			WantErr:      fc.WantErr,
			wantFail:     fc.WantFail, // set non exported field
			WantExitCode: fc.WantExitCode,
		}
	}
	return testCases
}

// Error overrides the embedded Error method to mock a test failure without
// actually failing when wantFail is equal to "test".
func (m *match) Error(args ...interface{}) {
	if m.wantFail != nil && *m.wantFail == "test" {
		return
	}
	m.T.Error(args...)
}

// FailNow overrides the embedded FailNow method to mock a test failure without
// actually failing when wantFail is equal to "test".
func (m *match) FailNow() {
	if m.wantFail != nil && *m.wantFail == "test" {
		*m.wantFail = "ok"
		return
	}
	m.T.FailNow()
}
