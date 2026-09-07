package yqlib

import (
	"strings"

	"go.yaml.in/yaml/v4"
)

// supplementaryPlaceholderStart and supplementaryPlaceholderEnd bound the
// Unicode Private Use Area used as stand-ins for supplementary-plane runes
// (see supplementaryRuneWorkaround). Both are single 3-byte UTF-8 code
// points that go.yaml.in/yaml/v4's emitter correctly treats as printable.
const (
	supplementaryPlaceholderStart = rune(0xE000)
	supplementaryPlaceholderEnd   = rune(0xF8FF)
)

// supplementaryRuneWorkaround masks runes outside the Basic Multilingual
// Plane (above U+FFFF) in multi-line string scalars before they reach
// go.yaml.in/yaml/v4's emitter, then restores the original bytes in the
// encoded output. It exists only to work around a bug in that vendored
// library: its scalar-printability scan checks byte values in a way that
// never recognises 4-byte UTF-8 sequences as printable, so any multi-line
// string containing e.g. an emoji outside the BMP is forced into a
// double-quoted, escaped rendering instead of the usual literal block style.
type supplementaryRuneWorkaround struct {
	originalByPlaceholder map[rune]rune
	next                  rune
}

func newSupplementaryRuneWorkaround() *supplementaryRuneWorkaround {
	return &supplementaryRuneWorkaround{
		originalByPlaceholder: map[rune]rune{},
		next:                  supplementaryPlaceholderStart,
	}
}

func (w *supplementaryRuneWorkaround) active() bool {
	return len(w.originalByPlaceholder) > 0
}

// remapNode walks a marshalled yaml.Node tree, replacing supplementary-plane
// runes in multi-line string scalars with placeholder runes.
func (w *supplementaryRuneWorkaround) remapNode(node *yaml.Node) {
	if node == nil {
		return
	}
	if node.Kind == yaml.ScalarNode && (node.Tag == "" || node.Tag == "!!str") {
		node.Value = w.remapValue(node.Value)
	}
	for _, child := range node.Content {
		w.remapNode(child)
	}
}

func (w *supplementaryRuneWorkaround) remapValue(value string) string {
	if !strings.ContainsRune(value, '\n') || !hasSupplementaryPlaneRune(value) {
		return value
	}

	placeholderByOriginal := map[rune]rune{}
	var builder strings.Builder
	for _, r := range value {
		if r <= 0xFFFF {
			builder.WriteRune(r)
			continue
		}
		placeholder, ok := placeholderByOriginal[r]
		if !ok {
			var found bool
			placeholder, found = w.reservePlaceholder(value, r)
			if !found {
				// Ran out of placeholders, or the text already legitimately
				// contains one of them: leave this rune as-is rather than
				// risk an incorrect restoration.
				builder.WriteRune(r)
				continue
			}
			placeholderByOriginal[r] = placeholder
		}
		builder.WriteRune(placeholder)
	}
	return builder.String()
}

func (w *supplementaryRuneWorkaround) reservePlaceholder(value string, original rune) (rune, bool) {
	for ; w.next <= supplementaryPlaceholderEnd; w.next++ {
		candidate := w.next
		if strings.ContainsRune(value, candidate) {
			continue
		}
		w.next++
		w.originalByPlaceholder[candidate] = original
		return candidate, true
	}
	return 0, false
}

// restore reverses the substitutions made by remapNode in the final encoded
// bytes, which are always safe: go.yaml.in/yaml/v4 only ever writes printable
// non-ASCII runes (which our placeholders are) as raw UTF-8, never escaped,
// regardless of the scalar style it ends up choosing.
func (w *supplementaryRuneWorkaround) restore(output []byte) []byte {
	if !w.active() {
		return output
	}
	result := string(output)
	for placeholder, original := range w.originalByPlaceholder {
		result = strings.ReplaceAll(result, string(placeholder), string(original))
	}
	return []byte(result)
}

func hasSupplementaryPlaneRune(s string) bool {
	for _, r := range s {
		if r > 0xFFFF {
			return true
		}
	}
	return false
}
