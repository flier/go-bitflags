package gen

import (
	"go/ast"
	"go/types"
)

type Package struct {
	Name  string
	Defs  map[*ast.Ident]types.Object
	Files []*File
}
