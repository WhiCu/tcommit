// Package template implements a template engine for TCommit.
// TODO: refactor
package template

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// placeholderRe matches patterns like {{.Key}} or {{.Key: choice1|choice2|@default}}
var placeholderRe = regexp.MustCompile(`\{\{\.(\w+)(?::\s*([^}]+))?\}\}`)

// Template holds template bytes, replacements, and an async result channel.
type Template struct {
	content      []byte
	replacements map[string]string
	rendered     chan []byte
}

// New loads the template from r into bytes and starts async parsing.
func New(r io.Reader, replacements map[string]string) *Template {
	data, err := io.ReadAll(r)
	if err != nil {
		panic("failed to read template: " + err.Error())
	}
	t := &Template{
		content:      data,
		replacements: replacements,
		rendered:     make(chan []byte, 1),
	}
	go t.parse()
	return t
}

// parse substitutes placeholders in the byte slice and sends result.
func (t *Template) parse() {
	// Use byte-based replacement to avoid unnecessary string conversions
	result := placeholderRe.ReplaceAllFunc(t.content, func(match []byte) []byte {
		parts := placeholderRe.FindSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		key := string(parts[1])
		var params string
		if len(parts) > 2 {
			params = string(parts[2])
		}

		allowed, def := parseParams(params)

		value, exists := t.replacements[key]
		if !exists {
			if def == "" {
				panic("missing required key: " + key)
			}
			value = def
		}

		if len(allowed) > 0 && !contains(allowed, value) {
			panic(fmt.Sprintf("invalid value '%s' for key '%s' (allowed: %v)", value, key, allowed))
		}
		return []byte(value)
	})
	t.rendered <- result
}

// Parse blocks until rendering completes and returns an io.Reader over bytes.
func (t *Template) Parse() io.Reader {
	data := <-t.rendered
	return bytes.NewReader(data)
}

// parseParams splits parameter string into allowed values and default.
func parseParams(params string) (allowed []string, def string) {
	for _, p := range strings.Split(params, "|") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "@") {
			def = p[1:]
		} else {
			allowed = append(allowed, p)
		}
	}
	return
}

// contains checks if val exists in slice.
func contains(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
