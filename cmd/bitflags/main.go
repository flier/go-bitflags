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
