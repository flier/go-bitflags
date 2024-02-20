# go-bitflags

[![Go](https://github.com/flier/go-bitflags/actions/workflows/go.yml/badge.svg)](https://github.com/flier/go-bitflags/actions/workflows/go.yml)
[![Report Card](https://goreportcard.com/badge/flier/go-bitflags)](https://goreportcard.com/report/flier/go-bitflags)
[![GoDoc](https://pkg.go.dev/badge/flier/go-bitflags.svg)](https://pkg.go.dev/flier/go-bitflags)

Bitflags is a tool to automate the creation of methods that satisfy the
`Bitflags` interface. Given the name of a (signed or unsigned) integer
type T that has constants defined, bitflags will create a new self-contained
Go source file implementing

    func (t T) Name() string

    func (t T) String() string

    func (t T) Contains(f T) bool

    func (t T) <Flag Name>() bool

The file is created in the same package and directory as the package that
defines T. It has helpful defaults designed for use with go generate.

## Installation

    # Go 1.18+
    go install github.com/flier/go-bitflags/cmd/bitflags@latest

    # Go version < 1.18
    go get -u github.com/flier/go-bitflags/cmd/bitflags@latest

## Usage

`Bitflags` works best with constants that are consecutive values such as
created using `iota`, but creates good code regardless. In the future it might
also provide custom support for constant sets that are bit patterns.

For example, given this snippet,

```go
package painkiller

type Pill int

const (
    Placebo Pill = 1 << iota
    Aspirin
    Ibuprofen
    Paracetamol
    Acetaminophen = Paracetamol
)
```

running this command

    bitflags -type=Pill

in the same directory will create the file pill_bitflags.go, in package
painkiller, containing a definition

``` go
// Name of a single flag bit, e.g. Placebo
func (Pill) Name() string

// All flag bits that have been set, such as Placebo|Ibuprofen
func (Pill) String() string

// Whether the flag bit is set
func (Pill) Contains(Pill) bool

func (Pill) Placebo() bool
func (Pill) Aspirin() bool
func (Pill) Ibuprofen() bool
func (Pill) Paracetamol() bool
func (Pill) Acetaminophen() bool
```

That method will translate the value of a Pill constant to the string
representation of the respective constant name, so that the call
`fmt.Print(painkiller.Aspirin|painkiller.Paracetamol)` will print the string
`Aspirin|Paracetamol`.

Typically this process would be run using go generate, like this:

    //go:generate bitflags --type=Pill

If multiple constants have the same value, the lexically first matching name
will be used (in the example, Acetaminophen will print as "Paracetamol").

With no arguments, it processes the package in the current directory.
Otherwise, the arguments must name a single directory holding a Go package
or a set of Go source files that represent a single Go package.

The `--type` flag accepts a comma-separated list of types so a single run can
generate methods for multiple types. The default output file is `t_bitflags.go`,
where t is the lower-cased name of the first type listed. It can be
overridden with the -output flag.

The `--line-comment` flag tells bitflags to generate the text of any line
comment, trimmed of leading spaces, instead of the constant name. For
instance, if the constants above had a Pill prefix, one could write

    PillAspirin // Aspirin

to suppress it in the output.

The `--trim-prefix` flag tell bitflags to trim the 'prefix' from the generated constant names.

The `--tag` flag tells bitflags the list of build tags to apply.

## Golang version

`bitflags` is currently compatible with golang version from 1.16+.

## Credits

The design of Bitflags was inspired by the [Stringer](https://pkg.go.dev/golang.org/x/tools/cmd/stringer) and [bitflags](https://github.com/bitflags/bitflags) crate, thanks to their great work.

## License

This project is licensed under either of Apache-2.0 License ([LICENSE-APACHE](LICENSE-APACHE)) or MIT license ([LICENSE-MIT](LICENSE-MIT)) at your option.

## Contribution

Unless you explicitly state otherwise, any contribution intentionally submitted
for inclusion in Futures by you, as defined in the Apache-2.0 license, shall be
dual licensed as above, without any additional terms or conditions.
