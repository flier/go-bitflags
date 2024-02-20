package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/flier/go-bitflags/internal/testenv"
)

// This file contains a test that compiles and runs each program in testdata
// after generating the string method for its type. The rule is that for testdata/x.go
// we run `bitflags -type X` and then compile and run the program. The resulting
// binary panics if the String method for X is not correct, including for error cases.

func TestMain(m *testing.M) {
	if os.Getenv("BITFLAGS_TEST_IS_BITFLAGS") != "" {
		main()
		os.Exit(0)
	}

	// Inform subprocesses that they should run the `cmd/bitflags` main instead of
	// running tests. It's a close approximation to building and running the real
	// command, and much less complicated and expensive to build and clean up.
	os.Setenv("BITFLAGS_TEST_IS_BITFLAGS", "1")

	flag.Parse()
	if testing.Verbose() {
		os.Setenv("GOPACKAGESDEBUG", "true")
	}

	os.Exit(m.Run())
}

func TestEndToEnd(t *testing.T) {
	testenv.NeedsTool(t, "go")

	executable := executablePath(t)
	// Read the testdata directory.
	fd, err := os.Open("testdata")
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()
	names, err := fd.Readdirnames(-1)
	if err != nil {
		t.Fatalf("Readdirnames: %s", err)
	}
	// Generate, compile, and run the test programs.
	for _, name := range names {
		if name == "typeparams" {
			// ignore the directory containing the tests with type params
			continue
		}
		if !strings.HasSuffix(name, ".go") {
			t.Errorf("%s is not a Go file", name)
			continue
		}
		if strings.HasPrefix(name, "tag_") || strings.HasPrefix(name, "vary_") {
			// This file is used for tag processing in TestTags or TestConstValueChange, below.
			continue
		}
		t.Run(name, func(t *testing.T) {
			if name == "cgo.go" {
				testenv.NeedsTool(t, "cgo")
			}
			bitflagsCompileAndRun(t, t.TempDir(), executable, typeName(name), name)
		})
	}
}

var exe struct {
	path string
	err  error
	once sync.Once
}

func executablePath(t *testing.T) string {
	testenv.NeedsExec(t)

	exe.once.Do(func() {
		exe.path, exe.err = os.Executable()
	})
	if exe.err != nil {
		t.Fatal(exe.err)
	}
	return exe.path
}

// a type name for stringer. use the last component of the file name with the .go
func typeName(fname string) string {
	// file names are known to be ascii and end .go
	base := path.Base(fname)
	return fmt.Sprintf("%c%s", base[0]+'A'-'a', base[1:len(base)-len(".go")])
}

// stringerCompileAndRun runs stringer for the named file and compiles and
// runs the target binary in directory dir. That binary will panic if the String method is incorrect.
func bitflagsCompileAndRun(t *testing.T, dir, executable, typeName, fileName string) {
	t.Logf("run: %s %s\n", fileName, typeName)
	source := filepath.Join(dir, path.Base(fileName))
	err := copyFile(source, filepath.Join("testdata", fileName))
	if err != nil {
		t.Fatalf("copying file to temporary directory: %s", err)
	}
	stringSource := filepath.Join(dir, typeName+"_string.go")
	// Run stringer in temporary directory.
	err = run(t, executable, "-type", typeName, "-output", stringSource, source)
	if err != nil {
		t.Fatal(err)
	}
	// Run the binary in the temporary directory.
	err = run(t, "go", "run", stringSource, source)
	if err != nil {
		t.Fatal(err)
	}
}

// copyFile copies the from file to the to file.
func copyFile(to, from string) error {
	toFd, err := os.Create(to)
	if err != nil {
		return err
	}
	defer toFd.Close()
	fromFd, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromFd.Close()
	_, err = io.Copy(toFd, fromFd)
	return err
}

// run runs a single command and returns an error if it does not succeed.
// os/exec should have this function, to be honest.
func run(t testing.TB, name string, arg ...string) error {
	t.Helper()
	return runInDir(t, ".", name, arg...)
}

// runInDir runs a single command in directory dir and returns an error if
// it does not succeed.
func runInDir(t testing.TB, dir, name string, arg ...string) error {
	t.Helper()
	cmd := testenv.Command(t, name, arg...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GO111MODULE=auto")
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		t.Logf("%s", out)
	}
	if err != nil {
		return fmt.Errorf("%v: %v", cmd, err)
	}
	return nil
}
