package test

import (
	"strconv"
	"strings"
)

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}

	_ = x[Placebo-1]
	_ = x[Aspirin-2]
	_ = x[Ibuprofen-4]
	_ = x[Paracetamol-8]
}

const (
	_Pill_name_0 = "PlaceboAspirin"
	_Pill_name_1 = "Ibuprofen"
	_Pill_name_2 = "Paracetamol"
)

var (
	_Pill_index_0 = [...]uint{0, 7, 14}
)

func (i Pill) Name() string {
	switch {
	case i == 4:
		return _Pill_name_1
	case i == 8:
		return _Pill_name_2
	default:
		return "Pill(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}

func (i Pill) Contains(f Pill) bool { return (i & f) == f }

func (i Pill) Placebo() bool { return i.Contains(Placebo) }

func (i Pill) Aspirin() bool { return i.Contains(Aspirin) }

func (i Pill) Ibuprofen() bool { return i.Contains(Ibuprofen) }

func (i Pill) Paracetamol() bool { return i.Contains(Paracetamol) }

func (i Pill) String() string {
	var b strings.Builder

	if i.Placebo() {
		if b.Len() > 0 {
			b.WriteByte('|')
		}
		b.WriteString("Placebo")
	}

	if i.Aspirin() {
		if b.Len() > 0 {
			b.WriteByte('|')
		}
		b.WriteString("Aspirin")
	}

	if i.Ibuprofen() {
		if b.Len() > 0 {
			b.WriteByte('|')
		}
		b.WriteString("Ibuprofen")
	}

	if i.Paracetamol() {
		if b.Len() > 0 {
			b.WriteByte('|')
		}
		b.WriteString("Paracetamol")
	}

	return b.String()
}
