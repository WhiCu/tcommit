package template

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testCase represents a set of test cases for template parsing and execution
type testCase struct {
	name         string
	template     string
	replacements map[string]string
	want         string
	parseErr     bool // Error expected during parsing
	executeErr   bool // Error expected during execution
}

var parseTests = []testCase{
	{
		name:         "Simple variable substitution",
		template:     "Hello {{.name}}!",
		replacements: map[string]string{"name": "World"},
		want:         "Hello World!",
		parseErr:     false,
		executeErr:   false,
	},
	{
		name:         "Multiple variables",
		template:     "{{.greeting}} {{.name}}!",
		replacements: map[string]string{"greeting": "Hello", "name": "World"},
		want:         "Hello World!",
		parseErr:     false,
		executeErr:   false,
	},
	{
		name:         "Variable with default value",
		template:     "Hello {{.name:@World}}!",
		replacements: map[string]string{},
		want:         "Hello World!",
		parseErr:     false,
		executeErr:   false,
	},
	{
		name:         "Variable with choices",
		template:     "Hello {{.greeting:Hi|Hello|@Hey}}!",
		replacements: map[string]string{"greeting": "Hello"},
		want:         "Hello Hello!",
		parseErr:     false,
		executeErr:   false,
	},
	{
		name:         "Invalid choice",
		template:     "Hello {{.greeting:Hi|Hello}}!",
		replacements: map[string]string{"greeting": "Hey"},
		want:         "",
		parseErr:     false,
		executeErr:   true,
	},
	{
		name:         "Missing variable without default",
		template:     "Hello {{.name}}!",
		replacements: map[string]string{},
		want:         "",
		parseErr:     false,
		executeErr:   true,
	},
	{
		name:         "Empty template",
		template:     "",
		replacements: map[string]string{},
		want:         "",
		parseErr:     false,
		executeErr:   false,
	},
	{
		name:         "No variables",
		template:     "Hello World!",
		replacements: map[string]string{},
		want:         "Hello World!",
		parseErr:     false,
		executeErr:   false,
	},
	{
		name:         "Unclosed template",
		template:     "Hello {{.name",
		replacements: map[string]string{"name": "World"},
		want:         "Hello {{.name",
		parseErr:     false,
		executeErr:   false,
	},
	{
		name:         "Multiple unclosed templates",
		template:     "{{.a {{.b {{.c",
		replacements: map[string]string{"a": "1", "b": "2", "c": "3"},
		want:         "{{.a {{.b {{.c",
		parseErr:     false,
		executeErr:   false,
	},
	{
		name:         "Invalid variable syntax",
		template:     "Hello {{name}}!",
		replacements: map[string]string{"name": "World"},
		want:         "",
		parseErr:     true,
		executeErr:   false,
	},
	{
		name:         "Invalid choice syntax",
		template:     "Hello {{.name:choice1|choice2|@default|invalid}}!",
		replacements: map[string]string{"name": "World"},
		want:         "",
		parseErr:     false,
		executeErr:   true,
	},
}

func TestParseString(t *testing.T) {
	for _, tc := range parseTests {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := ParseString(tc.template)
			if tc.parseErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			got, err := tmpl.Execute(ReplacerFuncFromMap(tc.replacements))
			if tc.executeErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestParseFromReader(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Valid template",
			input:   "Hello {{.name}}!",
			wantErr: false,
		},
		{
			name:    "Empty input",
			input:   "",
			wantErr: false,
		},
		{
			name:    "Invalid template",
			input:   "Hello {{name}}!",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.input)
			tmpl, err := Parse(reader)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, tmpl)
		})
	}
}

func TestExecuteTo(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		replacements map[string]string
		want         string
		wantErr      bool
	}{
		{
			name:         "Simple template",
			template:     "Hello {{.name}}!",
			replacements: map[string]string{"name": "World"},
			want:         "Hello World!",
			wantErr:      false,
		},
		{
			name:         "Template with error",
			template:     "Hello {{.name}}!",
			replacements: map[string]string{},
			want:         "",
			wantErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := ParseString(tc.template)
			require.NoError(t, err)

			var buf bytes.Buffer
			err = tmpl.ExecuteTo(&buf, ReplacerFuncFromMap(tc.replacements))
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, buf.String())
		})
	}
}

func TestConcurrentExecution(t *testing.T) {
	tmpl, err := ParseString("Hello {{.name}}!")
	require.NoError(t, err)

	replacer := ReplacerFuncFromMap(map[string]string{"name": "World"})
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			got, err := tmpl.Execute(replacer)
			require.NoError(t, err)
			assert.Equal(t, "Hello World!", got)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestConcurrentExecuteTo(t *testing.T) {
	tmpl, err := ParseString("Hello {{.name}}!")
	require.NoError(t, err)

	replacer := ReplacerFuncFromMap(map[string]string{"name": "World"})
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			var buf bytes.Buffer
			err := tmpl.ExecuteTo(&buf, replacer)
			require.NoError(t, err)
			assert.Equal(t, "Hello World!", buf.String())
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
