package bitflags

import "fmt"

type Bitflags interface {
	fmt.Stringer

	Name() string

	// Contains(f T) bool
}
