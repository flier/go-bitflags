package gen

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flier/go-bitflags/internal/testenv"
)

// Golden represents a test case.
type Golden struct {
	name        string
	trimPrefix  string
	lineComment bool
	typeName    string
	input       string // input; the package clause is provided when running the test.
	output      string // expected output.
}

//go:embed testdata/*.go
var testdata embed.FS

var golden = []Golden{
	{"pill", "", false, "Pill", "pill_in.go", "pill_out.go"},
}

const head = `package test

import (
	"strconv"
	"strings"
)

`

func TestGolden(t *testing.T) {
	testenv.NeedsTool(t, "go")

	dir := t.TempDir()
	for _, test := range golden {
		test := test
		t.Run(test.name, func(t *testing.T) {
			g := Generator{
				TrimPrefix:  test.trimPrefix,
				LineComment: test.lineComment,
				logf:        t.Logf,
			}

			file := test.name + ".go"
			absFile := filepath.Join(dir, file)

			buf, err := testdata.ReadFile("testdata/" + test.input)
			if err != nil {
				t.Fatal(err)
			}

			if err = os.WriteFile(absFile, buf, 0644); err != nil {
				t.Fatal(err)
			}

			g.ParsePackage([]string{absFile}, nil)

			if err = g.Generate(test.typeName); err != nil {
				t.Fatal(err)
			}

			got := head + strings.ReplaceAll(string(g.Format()), "\r\n", "\n")

			if buf, err = testdata.ReadFile("testdata/" + test.output); err != nil {
				t.Fatal(err)
			}

			if expected := strings.ReplaceAll(string(buf), "\r\n", "\n"); got != expected {
				t.Errorf("%s: got(%d)\n====\n%q\n====\nexpected(%d)\n====\n%q", test.name, len(got), got, len(expected), expected)
			}
		})
	}
}
