// Package template provides a simple template engine for string substitution.
// It supports basic variable substitution with optional choices and default values.
package template

import (
	"bytes"
	"io"
	"strings"
)

const (
	// Template markers
	openMarker  = "{{"
	closeMarker = "}}"

	// Variable markers
	varPrefix   = "."
	choiceSep   = ":"
	choiceDelim = "|"
	defPrefix   = "@"
)

// Node represents a part of the template: either text or a placeholder.
// It is an interface that defines how template parts are rendered.
type Node interface {
	// WriteTo writes the rendered node to w, using replacements from r.
	// It returns an error if the node cannot be rendered.
	WriteTo(w io.Writer, r Replacer) error
}

// Template holds parsed nodes and provides methods for template manipulation.
// It is the main type for working with templates.
type Template struct {
	Nodes []Node
}

// Parse reads the template from r and returns a Template.
// It reads the entire content of r into memory before parsing.
// Returns an error if reading fails.
//
// Syntax: {{.key}} or {{.key:choice1|choice2|@default}}
func Parse(r io.Reader) (*Template, error) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		return nil, err
	}
	return ParseString(buf.String())
}

// findNextTemplate finds the next template expression in the string.
// It searches for the pattern {{...}} starting from startPos.
// Returns the start and end positions of the template, and whether it was found.
func findNextTemplate(data string, startPos int) (start, end int, found bool) {
	openPos := strings.Index(data[startPos:], openMarker)
	if openPos < 0 {
		return 0, 0, false
	}
	openPos += startPos

	closePos := strings.Index(data[openPos+len(openMarker):], closeMarker)
	if closePos < 0 {
		return 0, 0, false
	}
	closePos += openPos + len(openMarker)

	return openPos, closePos + len(closeMarker), true
}

// ParseString parses a template string directly.
// It is more efficient than Parse when the input is already a string.
// Returns an error if the template syntax is invalid.
//
// Example:
//
//	tmpl, err := ParseString("Hello {{.name}}!")
func ParseString(data string) (*Template, error) {
	nodes := make([]Node, 0, len(data)/10) // Estimate initial capacity

	pos := 0
	for {
		// Find next template expression
		start, end, found := findNextTemplate(data, pos)
		if !found {
			// No more templates, add remaining text if any
			if len(data[pos:]) > 0 {
				nodes = append(nodes, &TextNode{Text: data[pos:]})
			}
			break
		}

		// Add text before template if any
		if start > pos {
			nodes = append(nodes, &TextNode{Text: data[pos:start]})
		}

		// Extract and parse template token
		token := data[start+len(openMarker) : end-len(closeMarker)]
		node, err := parseToken(token)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node)
		pos = end
	}

	return &Template{
		Nodes: nodes,
	}, nil
}

// parseToken parses a single template token into a Node.
// It handles both simple variables and variables with choices.
// Returns an error if the token syntax is invalid.
func parseToken(token string) (Node, error) {
	t := strings.TrimSpace(token)
	if !strings.HasPrefix(t, varPrefix) {
		return nil, NewInvalidTokenSyntaxError(token)
	}

	body := t[len(varPrefix):]
	var key, def string
	choices := make([]string, 0, 4) // Pre-allocate for common case
	hasDef := false

	if idx := strings.Index(body, choiceSep); idx >= 0 {
		key = strings.TrimSpace(body[:idx])
		rest := body[idx+len(choiceSep):]
		parts := strings.Split(rest, choiceDelim)

		for _, p := range parts {
			p = strings.TrimSpace(p)
			if strings.HasPrefix(p, defPrefix) {
				def = p[len(defPrefix):]
				hasDef = true
				// choices = append(choices, def)
			} else {
				choices = append(choices, p)
			}
		}
	} else {
		key = strings.TrimSpace(body)
	}

	return &VarNode{
		Key:     key,
		Choices: choices,
		Default: def,
		HasDef:  hasDef,
	}, nil
}

// Execute renders the template to a string using the provided Replacer.
// It processes all nodes in sequence and returns the final string.
// Returns an error if any node fails to render.
func (t *Template) Execute(r Replacer) (string, error) {
	var out strings.Builder
	out.Grow(len(t.Nodes) * 32) // Estimate average node size

	for _, node := range t.Nodes {
		if err := node.WriteTo(&out, r); err != nil {
			return "", err
		}
	}
	return out.String(), nil
}

// ExecuteTo writes the rendered template directly to the provided writer.
// It is more efficient than Execute when you want to write directly to a file or network connection.
// Returns an error if any node fails to render.
func (t *Template) ExecuteTo(w io.Writer, r Replacer) error {
	for _, node := range t.Nodes {
		if err := node.WriteTo(w, r); err != nil {
			return err
		}
	}
	return nil
}
