# golden

[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)
[![Coverage Status](http://codecov.io/github/larhun/golden/coverage.svg?branch=master)](http://codecov.io/github/larhun/golden?branch=master)
[![Build Status](https://travis-ci.org/larhun/golden.png?branch=master)](https://travis-ci.org/larhun/golden)
[![Go Report Card](https://goreportcard.com/badge/larhun/golden)](https://goreportcard.com/report/larhun/golden)
[![GoDoc](https://godoc.org/github.com/larhun/golden?status.svg)](https://godoc.org/github.com/larhun/golden)

----

Package golden provides functions for testing commands using smart validation
strings and gold master files.

A command must implement the `Runner` interface. It represents a black box
system with inputs (the argument list) and outputs (the standard and error
outputs, the panic and error messages, the exit code, and an optional file
output). A test case is a value of `Case` type. It represents a set of
input and output values. A test is performed using the `Test` function.

The following example tests a command named `"hello"` that should print `"Hello
World!"` to its standard output with an argument list equal to `[]string{"hello",
"World"}` (the first argument must be the command name).

```Go
    var command Runner = ... // the "hello" command

    func TestXxx(t *testing.T) {
        Test(t, command, []Case{{
            Name:       "output",                   // test case name
            Args:       []string{"hello", "World"}, // argument list
            WantStdout: "Hello World!",             // expected standard output
        })
    }
```

The command under test may be implemented using inner functions or may be
generated from an external `hello` program using the `Program` function:

```Go
    var command = Program("hello", nil)
```

The `Test` function supports smart validation strings and gold master files that
define very flexible matches with a very simple syntax (see the package
documentation for details). The `"..."` ellipsis is used to encode a partial
match:

```Go
    "Hello..."     // match "Hello" prefix
    "...World!"    // match "World!" suffix
    "...World..."  // match "World" substring
```

The `"^"` and `"$"` delimiters are used to encode a pattern matching:

```Go
    "^.*(Hello|World).*$" // match "Hello" or "World" substring
```

The `"golden"` term is used to encode a value stored by a gold master file:

```Go
    "golden.json" // match file testdata/golden/TestXxx-output.json
```

The gold master pattern is commonly used when testing complex output: the
expected string is saved to a file, the gold master, rather than to a validation
string. All the gold masters used by `TestXxx` are updated by running the test
with the `update` flag:

```Shell
    go test -run TestXxx -update
```

All the gold masters are updates by running:

```Shell
    go test -update
```

See the testing files for usage examples.
