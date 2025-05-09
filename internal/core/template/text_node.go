package template

import "io"

type TextNode struct {
	Text string
}

func (t *TextNode) WriteTo(w io.Writer, _ Replacer) error {
	_, err := io.WriteString(w, t.Text)
	return err
}
