package template

import (
	"io"
)

// VarNode holds a placeholder with optional choices and default.
type VarNode struct {
	Key     string
	Choices []string
	Default string
	HasDef  bool
}

// WriteTo writes the rendered node to w, using replacements.
func (v *VarNode) WriteTo(w io.Writer, r Replacer) error {
	val, found := r.Get(v.Key)
	if !found {
		if v.HasDef {
			val = v.Default
		} else {
			return NewNoReplacementError(v.Key)
		}
	}

	if len(v.Choices) > 0 && found {
		if !v.isValidChoice(val) {
			return NewInvalidValueError(val, v.Key, v.Choices)
		}
	}

	_, err := io.WriteString(w, val)
	return err
}

// isValidChoice checks if the given value is a valid choice.
func (v *VarNode) isValidChoice(val string) bool {
	for _, c := range v.Choices {
		if c == val {
			return true
		}
	}
	return false
}
