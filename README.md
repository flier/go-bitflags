# go-bitflags

Bitflags is a tool to automate the creation of methods that satisfy the
fmt.Stringer interface. Given the name of a (signed or unsigned) integer
type T that has constants defined, bitflags will create a new self-contained
Go source file implementing

    func (t T) String() string

The file is created in the same package and directory as the package that
defines T. It has helpful defaults designed for use with go generate.

## Installation

    # Go 1.16+
    go install github.com/flier/go-bitflags/cmd/bitflags@latest

    # Go version < 1.16
    go get -u github.com/flier/go-bitflags/cmd/bitflags@latest

## Usage

Bitflags works best with constants that are consecutive values such as
created using iota, but creates good code regardless. In the future it might
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
    func (Pill) Placebo() bool
    func (Pill) Aspirin() bool
    func (Pill) Ibuprofen() bool
    func (Pill) Paracetamol() bool
    func (Pill) Acetaminophen() bool
```

That method will translate the value of a Pill constant to the string
representation of the respective constant name, so that the call
fmt.Print(painkiller.Aspirin|painkiller.Paracetamol) will print the string
"Aspirin|Paracetamol".

Typically this process would be run using go generate, like this:

    //go:generate bitflags -type=Pill

If multiple constants have the same value, the lexically first matching name
will be used (in the example, Acetaminophen will print as "Paracetamol").

With no arguments, it processes the package in the current directory.
Otherwise, the arguments must name a single directory holding a Go package
or a set of Go source files that represent a single Go package.

The -type flag accepts a comma-separated list of types so a single run can
generate methods for multiple types. The default output file is t_string.go,
where t is the lower-cased name of the first type listed. It can be
overridden with the -output flag.

The -linecomment flag tells bitflags to generate the text of any line
comment, trimmed of leading spaces, instead of the constant name. For
instance, if the constants above had a Pill prefix, one could write

    PillAspirin // Aspirin

to suppress it in the output.

See https://github.com/flier/go-bitflags for more information and more examples.
