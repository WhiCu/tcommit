package template

import (
	"bytes"
	"strings"
	"testing"
)

var (
	// Common test data
	benchReplacements = map[string]string{
		"name":     "World",
		"greeting": "Hello",
		"default":  "Default",
	}

	// Template variations for benchmarking
	simpleTemplate     = "Hello {{.name}}!"
	complexTemplate    = "{{.greeting}} {{.name}}! How are you {{.name}}?"
	choiceTemplate     = "Hello {{.greeting:Hi|Hello|@Hey}} {{.name}}!"
	defaultTemplate    = "Hello {{.name:@World}}!"
	largeTemplate     = generateLargeTemplate(1000)
)

// generateLargeTemplate creates a large template with multiple variables
func generateLargeTemplate(size int) string {
	var b strings.Builder
	b.Grow(size * 50) // Estimate size

	for i := 0; i < size; i++ {
		b.WriteString("Line {{.greeting}} {{.name}}! ")
		if i%2 == 0 {
			b.WriteString("{{.name:@default}} ")
		}
		b.WriteString("\n")
	}
	return b.String()
}

func BenchmarkParseString(b *testing.B) {
	benchmarks := []struct {
		name     string
		template string
	}{
		{"Simple", simpleTemplate},
		{"Complex", complexTemplate},
		{"WithChoices", choiceTemplate},
		{"WithDefault", defaultTemplate},
		{"Large", largeTemplate},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, err := ParseString(bm.template)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecute(b *testing.B) {
	templates := []struct {
		name     string
		template string
	}{
		{"Simple", simpleTemplate},
		{"Complex", complexTemplate},
		{"WithChoices", choiceTemplate},
		{"WithDefault", defaultTemplate},
		{"Large", largeTemplate},
	}

	for _, tmpl := range templates {
		b.Run(tmpl.name, func(b *testing.B) {
			t, err := ParseString(tmpl.template)
			if err != nil {
				b.Fatal(err)
			}

			replacer := ReplacerFuncFromMap(benchReplacements)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := t.Execute(replacer)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteTo(b *testing.B) {
	templates := []struct {
		name     string
		template string
	}{
		{"Simple", simpleTemplate},
		{"Complex", complexTemplate},
		{"WithChoices", choiceTemplate},
		{"WithDefault", defaultTemplate},
		{"Large", largeTemplate},
	}

	for _, tmpl := range templates {
		b.Run(tmpl.name, func(b *testing.B) {
			t, err := ParseString(tmpl.template)
			if err != nil {
				b.Fatal(err)
			}

			replacer := ReplacerFuncFromMap(benchReplacements)
			buf := new(bytes.Buffer)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				buf.Reset()
				err := t.ExecuteTo(buf, replacer)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkConcurrentExecute(b *testing.B) {
	t, err := ParseString(largeTemplate)
	if err != nil {
		b.Fatal(err)
	}

	replacer := ReplacerFuncFromMap(benchReplacements)
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := t.Execute(replacer)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkConcurrentExecuteTo(b *testing.B) {
	t, err := ParseString(largeTemplate)
	if err != nil {
		b.Fatal(err)
	}

	replacer := ReplacerFuncFromMap(benchReplacements)
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		buf := new(bytes.Buffer)
		for pb.Next() {
			buf.Reset()
			err := t.ExecuteTo(buf, replacer)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
