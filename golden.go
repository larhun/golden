// Copyright (c) 2018 Larry Hunter <larhun.it@gmail.com>. All rights reserved.
//
// Use of this source code is governed by a BSD 3-Clause license that can be
// found in the LICENSE file.

/*
Package golden provides functions for testing commands using smart strings and
gold masters.

A command must implement the Runner interface. It represents a black box system
with inputs (the argument list) and outputs (the standard and error outputs, the
panic and error messages, the exit code, and an optional file output). A test
case is a value of Case type. It represents a set of input and output values. A
test is performed using the Test function.

The following example tests a command named "hello" that should print "Hello
World!" to its standard output with an argument list equal to []string{"hello",
"World"} (the first argument must be the command name).

    var command Runner = ... // the "hello" command

    func TestXxx(t *testing.T) {
        Test(t, command, []Case{{
            Name:       "output",                   // test case name
            Args:       []string{"hello", "World"}, // argument list
            WantStdout: "Hello World!",             // expected standard output
        })
    }

The command under test may be implemented using inner functions or may be
generated from an external hello program using the Program function:

    var command = Program("hello", nil)

The Test function supports smart validation strings and gold master files that
define very flexible matches. The gold master pattern is commonly used when
testing complex output: the expected string is saved to a file, the gold master,
rather than to a validation string.

The syntax of a smart validation string is very simple. A string equal to
"golden" with an optional extension encodes a match to the content of a gold
master file with the same extension. The file is stored in testdata/golden with
name derived from the test case name:

    "golden"         // match file testdata/golden/TestXxx-output
    "golden.json"    // match file testdata/golden/TestXxx-output.json

A string representing a valid regular expression delimited by the "^" and "$"
characters encodes a full pattern matching:

    "^value|error$" // match "value" or "error"

A string starting or ending with the "..." ellipsis encodes a partial match:

    "...value"    // match "value" suffix
    "value..."    // match "value" prefix
    "...value..." // match "value" substring

A string escaped by the equal symbol represents the substring after the symbol:

    "=value"    // match "value"
    "==value"   // match "=value"
    "=...value" // match "...value"
    "=^value$"  // match "^value$"

Any other string represents itself:

    "a...value"      // match "a...value"
    "^value|error"   // match "^value|error"
    "golden.file.go" // match "golden.file.go"

All the gold masters used by TestXxx are updated by running the test with the
update flag:

    go test -run TestXxx -update

All the gold masters are updates by running:

    go test -update

See the testing files for usage examples.

*/
package golden

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// goldenDir holds the directory of the gold master files.
var goldenDir = filepath.Join("testdata", "golden")

// update holds the updated flag.
var update = flag.Bool("update", false, "update gold masters in testdata/golden")

// Case represents a test case defined by a name and a set of input and output
// values.
type Case struct {
	// Name holds the case name, which is used to uniquely identify the test
	// case. A trailing number may be added for disambiguation.
	Name string

	// Args holds the argument list that is passed to the command under test.
	Args []string

	// WantFile contains the name of the file that should be written by the
	// command under test. If exists, the file is removed before running the
	// test. The expected content is stored by a gold master file with the same
	// extension. The TmpFiles helper function should be used to test files in a
	// temporary directory.
	WantFile string

	// WantStdout and WantStderr hold a smart validation string for the expected
	// standard and error output, respectively.
	WantStdout string
	WantStderr string

	// WantPanic and WantErr hold a smart validation string for the expected
	// panic and error message, respectively.
	WantPanic string
	WantErr   string

	// wantFail is used for inner testing. It holds a smart validation string
	// for the expected error message from a test that must fail. If matches,
	// the test passes rather than failing. The magic word "test" is used to
	// mock a test failure without actually failing.
	wantFail *string

	// WantExitCode holds the expected exit code.
	WantExitCode int
}

// TmpFiles adds to all the named files the full path to a new temporary
// directory and returns a cleanup function. It must be used within a defer
// statement:
//
//     defer TmpFiles(t, name1, name2, ...)()
//
func TmpFiles(t *testing.T, names ...*string) func() {
	dir, err := ioutil.TempDir("", "go-golden")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		*name = filepath.Join(dir, *name)
	}
	return func() {
		os.RemoveAll(dir)
	}
}

// Test tests the specified command by running a subtest for each listed case
// with the provided argument list. If the command outputs do not match the
// expected ones, the subtest signals a failure and reports an error for each
// invalid output.
func Test(t *testing.T, command Runner, testCases []Case) {
	t.Helper()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Helper() // TODO: make Helper working for subtests: issue #24128

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			command.SetStdout(stdout)
			command.SetStderr(stderr)

			m := newMatch(t, tc.wantFail)

			if tc.WantFile != "" {
				if !m.removeFile(tc.WantFile) {
					tc.WantFile = "" // stop testing File match
				}
			}

			var gotErr string
			gotPanic := m.run(func() {
				if err := command.Run(tc.Args); err != nil {
					gotErr = err.Error()
				}
			})

			if tc.WantFile != "" {
				if gotFile, ext, ok := m.getFile(tc.WantFile); ok {
					m.match("File golden"+ext, gotFile, "golden"+ext)
				}
			}

			m.match("WantStdout", stdout.String(), tc.WantStdout)
			m.match("WantStderr", stderr.String(), tc.WantStderr)
			m.match("WantPanic", gotPanic, tc.WantPanic)
			m.match("WantErr", gotErr, tc.WantErr)
			m.equal("WantExitCode", command.ExitCode(), tc.WantExitCode)

			m.done()
		})
	}
}

// Runner is the interface implemented by a command. It represents a black box
// system with inputs (the argument list) and outputs (the standard and error
// outputs, the panic and error messages, the exit code, and an optional file
// output).
type Runner interface {
	// Run executes the command with the specified argument list. The first
	// argument must be the command name.
	Run(args []string) error

	// SetStdout and SetStderr set the command's standard and error outputs.
	SetStdout(w io.Writer)
	SetStderr(w io.Writer)

	// ExitCode returns the exit code of the last run: 0 before the first run
	// and after a successful run; 2 if the command name is missing or invalid;
	// -1 if the code cannot be recovered.
	ExitCode() int
}

// program implements a Runner for an external program.
type program struct {
	name     string    // program name
	env      []string  // process environment
	stdout   io.Writer // standard output
	stderr   io.Writer // standard error
	exitCode int       // exit code
}

func (p *program) Run(args []string) error {
	p.exitCode = 2

	if len(args) == 0 {
		return errors.New("missing program name")
	}
	if args[0] != p.name {
		return errors.New("invalid program name: " + args[0])
	}

	p.exitCode = 0

	cmd := exec.Command(p.name, args[1:]...)
	cmd.Stdout = p.stdout
	cmd.Stderr = p.stderr
	cmd.Env = p.env

	err := cmd.Run()
	if err != nil {
		type status interface {
			ExitStatus() int
		}

		p.exitCode = -1

		if s, ok := cmd.ProcessState.Sys().(status); ok {
			p.exitCode = s.ExitStatus()
		}
	}

	return err
}

func (p *program) SetStdout(w io.Writer) { p.stdout = w }
func (p *program) SetStderr(w io.Writer) { p.stderr = w }
func (p *program) ExitCode() int         { return p.exitCode }

// Program returns a Runner for the named program and the specified process
// environment. See exec.Command and exec.Cmd.Env for valid name and env values.
func Program(name string, env []string) Runner {
	return &program{
		name: name,
		env:  append(os.Environ(), env...),
	}
}
