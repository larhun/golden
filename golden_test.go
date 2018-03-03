// Copyright (c) 2018 Larry Hunter <larhun.it@gmail.com>.
// All rights reserved.
//
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package golden_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/larhun/golden"
)

// ptrTo returns the pointer of a string
func ptrTo(s string) *string {
	return &s
}

// mock represents a simple mocking function of an external hello program.
func mock() int {
	if len(os.Args) == 1 {
		panic("missing command")
	}

	value := strings.Join(os.Args[2:], "\n")

	switch os.Args[1] {
	case "stdout":
		fmt.Fprint(os.Stdout, value)
		return 0
	case "stderr":
		fmt.Fprint(os.Stderr, value)
		return 1
	}

	panic("invalid command name: " + os.Args[1])
}

// echo represents a simple echo command that implements the Runner interface.
type echo struct {
	stdout   io.Writer
	stderr   io.Writer
	exitCode int
}

func (e *echo) Run(args []string) error {
	e.exitCode = 2
	if len(args) == 0 {
		return errors.New("missing command name")
	}
	if args[0] != "echo" {
		return errors.New("invalid command name: " + args[0])
	}
	if len(args) == 1 || args[1] == "file" && len(args) != 3 {
		return errors.New("bad number of arguments")
	}

	value := strings.Join(args[2:], "\n")

	e.exitCode = 0
	switch args[1] {

	case "stdout":
		e.stdout.Write([]byte(value))

	case "stderr":
		e.exitCode = 1
		e.stderr.Write([]byte(value))

	case "panic":
		e.exitCode = 2
		panic(value)

	case "err":
		e.exitCode = 3
		return errors.New(value)

	case "exit":
		e.exitCode = 4

	case "file":
		if err := ioutil.WriteFile(args[2], []byte(filepath.Base(args[2])), 0666); err != nil {
			panic(err)
		}
	}

	return nil
}

func (e *echo) SetStdout(w io.Writer) { e.stdout = w }
func (e *echo) SetStderr(w io.Writer) { e.stderr = w }
func (e echo) ExitCode() int          { return e.exitCode }

func TestMain(m *testing.M) {
	if _, ok := os.LookupEnv("GOLDEN_TEST_MOCK"); ok {
		os.Exit(mock())
	}

	os.Exit(m.Run())
}

func TestEqual(t *testing.T) {
	Test(t, new(echo), []Case{
		{
			Args:         []string{"echo", "stdout", "value"},
			WantStdout:   "value",
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stderr", "value"},
			WantStderr:   "value",
			WantExitCode: 1,
		}, {
			Args:         []string{"echo", "panic", "value"},
			WantPanic:    "value",
			WantExitCode: 2,
		}, {
			Args:         []string{"echo", "err", "value"},
			WantErr:      "value",
			WantExitCode: 3,
		},
	})
}

func TestEqualBadSyntax(t *testing.T) {
	Test(t, new(echo), []Case{
		{
			Args:         []string{"echo", "stdout", "^value"},
			WantStdout:   "^value",
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stdout", "value$"},
			WantStdout:   "value$",
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stdout", "^(value$"},
			WantStdout:   "^(value$",
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stdout", "golden.ext.ext"},
			WantStdout:   "golden.ext.ext",
			WantExitCode: 0,
		},
	})
}

func TestGolden(t *testing.T) {
	Test(t, new(echo), []Case{
		{
			Args:         []string{"echo", "stdout", "stdout"},
			WantStdout:   "golden.stdout",
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stderr", "stderr"},
			WantStderr:   "golden.stderr",
			WantExitCode: 1,
		}, {
			Args:         []string{"echo", "panic", "panic"},
			WantPanic:    "golden.panic",
			WantExitCode: 2,
		}, {
			Args:         []string{"echo", "err", "err"},
			WantErr:      "golden.err",
			WantExitCode: 3,
		},
	})
}

func TestPattern(t *testing.T) {
	Test(t, new(echo), []Case{
		{
			Args:         []string{"echo", "stdout", "value"},
			WantStdout:   "^value$",
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stderr", "value"},
			WantStderr:   "^.*ue$",
			WantExitCode: 1,
		}, {
			Args:         []string{"echo", "panic", "value"},
			WantPanic:    "^va.*$",
			WantExitCode: 2,
		}, {
			Args:         []string{"echo", "err", "value"},
			WantErr:      "^.*l.*$",
			WantExitCode: 3,
		},
	})
}

func TestEllipsis(t *testing.T) {
	Test(t, new(echo), []Case{
		{
			Args:         []string{"echo", "stdout", "value"},
			WantStdout:   "...ue",
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stderr", "value"},
			WantStderr:   "...l...",
			WantExitCode: 1,
		}, {
			Args:         []string{"echo", "panic", "value"},
			WantPanic:    "va...",
			WantExitCode: 2,
		}, {
			Args:         []string{"echo", "err", "value"},
			WantErr:      "...value...",
			WantExitCode: 3,
		},
	})
}

func TestEscaped(t *testing.T) {
	Test(t, new(echo), []Case{
		{
			Args:         []string{"echo", "stdout", "value"},
			WantStdout:   "=value",
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stderr", "value"},
			WantStderr:   "=value",
			WantExitCode: 1,
		}, {
			Args:         []string{"echo", "panic", "value"},
			WantPanic:    "=value",
			WantExitCode: 2,
		}, {
			Args:         []string{"echo", "err", "value"},
			WantErr:      "=value",
			WantExitCode: 3,
		},
	})
}

func TestFile(t *testing.T) {
	found, missing := "found", "missing"
	defer TmpFiles(t, &found, &missing)()

	if err := ioutil.WriteFile(found, nil, 0600); err != nil {
		t.Fatal("cannot write temporary file: " + found)
	}

	Test(t, new(echo), []Case{
		{
			Args:         []string{"echo", "file", found},
			WantFile:     found,
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "file", missing},
			WantFile:     missing,
			WantExitCode: 0,
		},
	})
}

func TestExternal(t *testing.T) {
	Test(t, Program("go", nil), []Case{
		// go version outputs
		{
			Args:         []string{"go", "version"},
			WantStdout:   "go version...",
			WantExitCode: 0,
		}, {
			Args:         []string{"go", "version", "-help"},
			WantStderr:   "...usage: version...",
			WantErr:      "exit status 2",
			WantExitCode: 2,
		},
		// calling errors
		{
			Args:         []string{},
			WantErr:      "missing program name",
			WantExitCode: 2,
		}, {
			Args:         []string{"version"},
			WantErr:      "invalid program name: version",
			WantExitCode: 2,
		},
	})
}

func TestMock(t *testing.T) {
	name := os.Args[0]
	Test(t, Program(name, []string{"GOLDEN_TEST_MOCK="}), []Case{
		{
			Args:         []string{name, "stdout"},
			WantExitCode: 0,
		}, {
			Args:         []string{name, "stderr"},
			WantErr:      "exit status 1",
			WantExitCode: 1,
		}, {
			Args:         []string{name},
			WantStderr:   "...panic: missing command...",
			WantErr:      "exit status 2",
			WantExitCode: 2,
		}, {
			Args:         []string{name, "command"},
			WantStderr:   "...panic: invalid command name: ...",
			WantErr:      "exit status 2",
			WantExitCode: 2,
		},
	})
}

func TestErrors(t *testing.T) {
	missing, file, dir := "missing", "file", "dir"
	defer TmpFiles(t, &missing, &file, &dir)()

	if err := ioutil.WriteFile(file, nil, 0600); err != nil {
		t.Fatal("cannot write temporary file: " + file)
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal("cannot write temporary directory: " + dir)
	}

	Test(t, new(echo), ToCase([]FailCase{
		// bad arguments
		{
			Args:         nil,
			WantFail:     ptrTo("WantErr match error:\nwant an empty string, got: missing command name"),
			WantExitCode: 2,
		}, {
			Args:         []string{"echo"},
			WantFail:     ptrTo("WantErr match error:\nwant an empty string, got: bad number of arguments"),
			WantExitCode: 2,
		},
		// bad match
		{
			Args:         []string{"echo", "stdout", "value"},
			WantFail:     ptrTo("WantStdout match error:\nwant an empty string, got: value"),
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stderr", "value"},
			WantFail:     ptrTo("WantStderr match error:\nwant an empty string, got: value"),
			WantExitCode: 1,
		}, {
			Args:         []string{"echo", "panic", "value"},
			WantFail:     ptrTo("WantPanic match error:\nwant an empty string, got: value"),
			WantExitCode: 2,
		}, {
			Args:         []string{"echo", "err", "value"},
			WantFail:     ptrTo("WantErr match error:\nwant an empty string, got: value"),
			WantExitCode: 3,
		}, {
			Args:         []string{"echo", "exit"},
			WantFail:     ptrTo("WantExitCode match error:\ngot: 4, want: 0"),
			WantExitCode: 0,
		},
		// bad golden file
		{
			Args:         []string{"echo", "stdout", "value"},
			WantStdout:   "golden.ext",
			WantFail:     ptrTo("WantStdout golden.ext read error:\n..."),
			WantExitCode: 0,
		},
		// bad file
		{
			Args:         []string{"echo", "file", file},
			WantFile:     missing,
			WantFail:     ptrTo("File read error:\n..."),
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stdout"},
			WantFile:     dir,
			WantFail:     ptrTo("File mode error:\nexpected regular file"),
			WantExitCode: 0,
		},
	}))
}

func TestFormat(t *testing.T) {
	Test(t, new(echo), ToCase([]FailCase{
		// single-line error format with one empty string
		{
			Args:         []string{"echo", "stdout", ""},
			WantStdout:   "value\n",
			WantFail:     ptrTo("WantStdout match error:\ngot an empty string, want: value"),
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stdout", "value\n"},
			WantStdout:   "",
			WantFail:     ptrTo("WantStdout match error:\nwant an empty string, got: value"),
			WantExitCode: 0,
		},
		// multi-line error format with one empty string
		{
			Args:         []string{"echo", "stdout", ""},
			WantStdout:   "value\nvalue",
			WantFail:     ptrTo("WantStdout match error:\ngot an empty string, want:\n    value\n    value"),
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stdout", "value\nvalue"},
			WantStdout:   "",
			WantFail:     ptrTo("WantStdout match error:\nwant an empty string, got:\n    value\n    value"),
			WantExitCode: 0,
		},
		// single-line error format for both strings (skip ending newline)
		{
			Args:         []string{"echo", "stdout", "value"},
			WantStdout:   "value\n",
			WantFail:     ptrTo("WantStdout match error:\ngot: value\nwant: value"),
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stdout", "value\n"},
			WantStdout:   "value",
			WantFail:     ptrTo("WantStdout match error:\ngot: value\nwant: value"),
			WantExitCode: 0,
		},
		// multi-line error format for both strings
		{
			Args:         []string{"echo", "stdout", "value"},
			WantStdout:   "value\nvalue",
			WantFail:     ptrTo("WantStdout match error:\ngot:\n    value\nwant:\n    value\n    value"),
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stdout", "value\nvalue"},
			WantStdout:   "value",
			WantFail:     ptrTo("WantStdout match error:\ngot:\n    value\n    value\nwant:\n    value"),
			WantExitCode: 0,
		},
		// smart string error format
		{
			Args:         []string{"echo", "stdout"},
			WantStdout:   "^value$",
			WantFail:     ptrTo("WantStdout pattern match error:\ngot an empty string, want: ^value$"),
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stdout"},
			WantStdout:   "...value...",
			WantFail:     ptrTo("WantStdout substring match error:\ngot an empty string, want: ...value..."),
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stdout"},
			WantStdout:   "value...",
			WantFail:     ptrTo("WantStdout prefix match error:\ngot an empty string, want: value..."),
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stdout"},
			WantStdout:   "...value",
			WantFail:     ptrTo("WantStdout suffix match error:\ngot an empty string, want: ...value"),
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stdout"},
			WantStdout:   "=value",
			WantFail:     ptrTo("WantStdout escaped match error:\ngot an empty string, want: =value"),
			WantExitCode: 0,
		},
	}))
}

func TestInner(t *testing.T) {
	pass, fail := "test", "test"

	Test(t, new(echo), ToCase([]FailCase{
		{
			Args:         []string{"echo", "stdout", "value"},
			WantStdout:   "value",
			WantFail:     &pass,
			WantExitCode: 0,
		}, {
			Args:         []string{"echo", "stdout", "value"},
			WantFail:     &fail,
			WantExitCode: 0,
		},
	}))

	if pass != "ok" {
		t.Fatal("expected ok from inner pass test")
	}

	if fail != "ok" {
		t.Fatal("expected ok from inner fail test")
	}
}
