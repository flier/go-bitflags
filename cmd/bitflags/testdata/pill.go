package main

import "fmt"

type Pill int

const (
	Placebo Pill = 1 << iota
	Aspirin
	Ibuprofen
	Paracetamol
	Acetaminophen = Paracetamol
)

func main() {
	ck(Placebo, "Placebo")
	ck(Aspirin, "Aspirin")
	ck(Ibuprofen, "Ibuprofen")
	ck(Paracetamol, "Paracetamol")
	ck(Acetaminophen, "Paracetamol")
	ck(Placebo|Aspirin, "Placebo|Aspirin")
	ck(127, "Placebo|Aspirin|Ibuprofen|Paracetamol")
}

func ck(pill Pill, expected string) {
	if got := fmt.Sprint(pill); got != expected {
		panic("pill.go: \n\tgot: " + got + "\n\texpected:" + expected)
	}
}
