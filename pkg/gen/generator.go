package gen

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

//go:embed templates/*
var content embed.FS

var templates = template.Must(template.New("templates").Funcs(template.FuncMap{
	"lower": strings.ToLower,
	"upper": strings.ToUpper,
	"last":  func(values []Value) Value { return values[len(values)-1] },
}).ParseFS(content, "templates/*.go.tmpl"))

type Generator struct {
	buf  bytes.Buffer
	pkg  *Package                                 // Package we are scanning.
	logf func(format string, args ...interface{}) // test logging hook; nil when not testing

	TrimPrefix  string
	LineComment bool
}

func New(trimPrefix string, lineComment bool) *Generator {
	return &Generator{
		TrimPrefix:  trimPrefix,
		LineComment: lineComment,
	}
}

var (
	ErrTooManyPackages = errors.New("too many package")
)

// ParsePackage analyzes the single package constructed from the patterns and tags.
func (g *Generator) ParsePackage(patterns []string, tags []string) (err error) {
	cfg := &packages.Config{
		Mode:       packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		Tests:      false,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
		Logf:       g.logf,
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		err = fmt.Errorf("load package, %w", err)
		return
	}

	if len(pkgs) != 1 {
		err = fmt.Errorf("%d packages matching %v, %w", len(pkgs), strings.Join(patterns, " "), ErrTooManyPackages)
		return
	}

	g.addPackage(pkgs[0])

	return
}

// addPackage adds a type checked Package and its syntax files to the generator.
func (g *Generator) addPackage(pkg *packages.Package) {
	g.pkg = &Package{
		Name:  pkg.Name,
		Defs:  pkg.TypesInfo.Defs,
		Files: make([]*File, len(pkg.Syntax)),
	}

	for i, file := range pkg.Syntax {
		g.pkg.Files[i] = &File{
			File:        file,
			Package:     g.pkg,
			TrimPrefix:  g.TrimPrefix,
			LineComment: g.LineComment,
		}
	}
}

func (g *Generator) GenerateHeader() error {
	return templates.Lookup("header.go.tmpl").Execute(&g.buf, map[string]interface{}{
		"CmdLine": strings.Join(append([]string{filepath.Base(os.Args[0])}, os.Args[1:]...), " "),
		"Package": g.pkg,
	})
}

// generate produces the String method for the named type.
func (g *Generator) Generate(typeName string) (err error) {
	values := make([]Value, 0, 100)
	for _, file := range g.pkg.Files {
		// Set the state for this run of the walker.
		file.TypeName = typeName
		file.Values = nil
		if file.File != nil {
			ast.Inspect(file.File, file.GenDecl)
			values = append(values, file.Values...)
		}
	}

	if len(values) == 0 {
		err = fmt.Errorf("no values defined for type %s", typeName)
		return
	}

	// Generate code that will fail if the constants change value.
	if err = templates.Lookup("validate.go.tmpl").Execute(&g.buf, values); err != nil {
		return
	}

	if runs := splitIntoRuns(values); len(runs) <= 10 {
		if err = g.declareIndexAndNameVars(runs, typeName); err != nil {
			return
		}
	} else {
		if err = g.declareMapAndNameVars(values, typeName); err != nil {
			return
		}
	}

	if err = templates.Lookup("props.go.tmpl").Execute(&g.buf, map[string]interface{}{
		"Type":   typeName,
		"Values": values,
	}); err != nil {
		return
	}

	return
}

// splitIntoRuns breaks the values into runs of contiguous sequences.
// For example, given 1,2,3,5,6,7 it returns {1,2,3},{5,6,7}.
// The input slice is known to be non-empty.
func splitIntoRuns(v []Value) [][]Value {
	// We use stable sort so the lexically first name is chosen for equal elements.
	values := make([]Value, len(v))
	copy(values, v)
	sort.Stable(byValue(values))
	// Remove duplicates. Stable sort has put the one we want to print first,
	// so use that one. The String method won't care about which named constant
	// was the argument, so the first name for the given value is the only one to keep.
	// We need to do this because identical values would cause the switch or map
	// to fail to compile.
	j := 1
	for i := 1; i < len(values); i++ {
		if values[i].Value != values[i-1].Value {
			values[j] = values[i]
			j++
		}
	}
	values = values[:j]
	runs := make([][]Value, 0, 10)
	for len(values) > 0 {
		// One contiguous sequence per outer loop.
		i := 1
		for i < len(values) && values[i].Value == values[i-1].Value+1 {
			i++
		}
		runs = append(runs, values[:i])
		values = values[i:]
	}
	return runs
}

func (g *Generator) declareIndexAndNameVars(runs [][]Value, typeName string) error {
	var indexes [][]int
	for i, run := range runs {
		if len(run) > 1 {
			if indexes == nil {
				indexes = make([][]int, len(runs))
			}

			indexes[i] = make([]int, len(run))
			off := 0
			for j, v := range run {
				off += len(v.OriginalName)
				indexes[i][j] = off
			}
		}
	}

	return templates.Lookup("index_and_name.go.tmpl").Execute(&g.buf, map[string]interface{}{
		"Runs":    runs,
		"Indexes": indexes,
		"Type":    typeName,
	})
}

func (g *Generator) declareMapAndNameVars(values []Value, typeName string) error {
	type value struct {
		Value
		Start, End int
	}

	_values := make([]value, len(values))
	off := 0

	for i, v := range values {
		_values[i] = value{v, off, off + len(v.OriginalName)}
		off += len(v.OriginalName)
	}

	return templates.Lookup("map_and_name.go.tmpl").Execute(&g.buf, map[string]interface{}{
		"Values": _values,
		"Type":   typeName,
	})
}

func (g *Generator) Format() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		return g.buf.Bytes()
	}
	return src
}
