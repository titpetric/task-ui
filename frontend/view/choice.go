package view

// Choice is an option of a row's select: its value, its label and
// whether the row's rule holds it.
type Choice struct {
	Value, Label string
	Selected     bool
}
