package test

type Pill int

const (
	Placebo Pill = 1 << iota
	Aspirin
	Ibuprofen
	Paracetamol
	Acetaminophen = Paracetamol
)
