// Bitflags is a tool to automate the creation of methods that satisfy the fmt.Stringer
// interface. Given the name of a (signed or unsigned) integer type T that has constants
// defined, bitflags will create a new self-contained Go source file implementing
//	func (t T) String() string
// The file is created in the same package and directory as the package that defines T.
// It has helpful defaults designed for use with go generate.
//
// Bitflags works best with constants that are consecutive values such as created using iota,
// but creates good code regardless. In the future it might also provide custom support for
// constant sets that are bit patterns.
//
// For example, given this snippet,
//
//	package painkiller
//
//	type Pill int
//
//	const (
//		Placebo Pill = 1 << iota
//		Aspirin
//		Ibuprofen
//		Paracetamol
//		Acetaminophen = Paracetamol
//	)
//
// running this command
//
//	bitflags -type=Pill
//
// in the same directory will create the file pill_bitflags.go, in package painkiller,
// containing a definition of
//
//	// Name of a single flag bit, e.g. Placebo
//	func (Pill) Name() string
//
//	// All flag bits that have been set, such as Placebo|Ibuprofen
//	func (Pill) String() string
//
//	// Whether the flag bit is set
//	func (Pill) Placebo() bool
//	func (Pill) Aspirin() bool
//	func (Pill) Ibuprofen() bool
//	func (Pill) Paracetamol() bool
//	func (Pill) Acetaminophen() bool
//
// That method will translate the value of a Pill constant to the string representation
// of the respective constant name, so that the call fmt.Print(painkiller.Aspirin|painkiller.Paracetamol) will
// print the string "Aspirin|Paracetamol".
//
// Typically this process would be run using go generate, like this:
//
//	//go:generate bitflags -type=Pill
//
// If multiple constants have the same value, the lexically first matching name will
// be used (in the example, Acetaminophen will print as "Paracetamol").
//
// With no arguments, it processes the package in the current directory.
// Otherwise, the arguments must name a single directory holding a Go package
// or a set of Go source files that represent a single Go package.
//
// The -type flag accepts a comma-separated list of types so a single run can
// generate methods for multiple types. The default output file is t_string.go,
// where t is the lower-cased name of the first type listed. It can be overridden
// with the -output flag.
//
// The -linecomment flag tells bitflags to generate the text of any line comment, trimmed
// of leading spaces, instead of the constant name. For instance, if the constants above had a
// Pill prefix, one could write
//
//	PillAspirin // Aspirin
//
// to suppress it in the output.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jpillora/opts"

	"github.com/flier/go-bitflags/pkg/gen"
)

type config struct {
	Types       []string `opts:"help=list of type names"`
	Output      string   `opts:"help=output file name; default srcdir/<type>_bitflags.go"`
	TrimPrefix  string   `opts:"help=trim the 'prefix' from the generated constant names"`
	LineComment bool     `opts:"help=use line comment text as printed text when present"`
	Tags        []string `opts:"help=list of build tags to apply"`
	Files       []string `opts:"mode=arg,help=package directory or a list of files"`
}

const (
	prog    = "bitflags"
	author  = "Flier Lu <flier.lu@gmail.com>"
	repo    = "https://github.com/flier/go-bitflags"
	summary = `
    {{ .Name }} [flags] -type T [directory]
    {{ .Name }} [flags] -type T files... # Must be a single package

For more information, see:

    https://github.com/flier/go-bitflags/cmd/bitflags`
)

func main() {
	var b strings.Builder
	template.Must(template.New("summary").Parse(summary)).Execute(&b, map[string]interface{}{"Name": prog})
	summary := strings.TrimRight(b.String(), "\n")

	c := config{}
	opts.New(&c).Author(author).Summary(summary).Repo(repo).Parse()

	g := gen.New(c.TrimPrefix, c.LineComment)

	if len(c.Files) == 0 {
		c.Files = []string{"."}
	}

	if err := g.ParsePackage(c.Files, c.Tags); err != nil {
		log.Fatalf("fail to parse package, %v", err)
	}

	g.GenerateHeader()

	// Run generate for each type.
	for _, s := range c.Types {
		for _, typeName := range strings.Split(s, ",") {
			if err := g.Generate(typeName); err != nil {
				log.Fatalf("fail to generate %s, %v", typeName, err)
			}
		}
	}

	src := g.Format()

	// Write to file.
	if c.Output == "-" {
		fmt.Print(string(src))
	} else {
		outputName := c.Output
		if outputName == "" {
			var dir string

			if len(c.Files) == 1 && isDirectory(c.Files[0]) {
				dir = c.Files[0]
			} else {
				if len(c.Tags) != 0 {
					log.Fatal("-tags option applies only to directories, not when files are specified")
				}
				dir = filepath.Dir(c.Files[0])
			}

			baseName := strings.ToLower(fmt.Sprintf("%s_bitflags.go", c.Types[0]))
			outputName = filepath.Join(dir, baseName)
		}

		if err := os.WriteFile(outputName, src, 0644); err != nil {
			log.Fatalf("writing output: %s", err)
		}
	}
}

func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}
