package beszel

import (
	"encoding/json"
)

// Pair is two numbers a record keeps as an array, as sent and received.
// Anything else decodes to zero: a field an agent fills differently
// costs that field alone, and the record decodes.
type Pair [2]float64

// UnmarshalJSON reads an array of two numbers and ignores the rest.
func (p *Pair) UnmarshalJSON(data []byte) error {
	var pair [2]float64
	if json.Unmarshal(data, &pair) == nil {
		*p = pair
	}
	return nil
}
