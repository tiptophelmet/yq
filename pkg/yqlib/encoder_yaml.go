package yqlib

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"go.yaml.in/yaml/v4"
)

type yamlEncoder struct {
	prefs YamlPreferences
}

func NewYamlEncoder(prefs YamlPreferences) Encoder {
	return &yamlEncoder{prefs}
}

func (ye *yamlEncoder) CanHandleAliases() bool {
	return true
}

func (ye *yamlEncoder) PrintDocumentSeparator(writer io.Writer) error {
	return PrintYAMLDocumentSeparator(writer, ye.prefs.PrintDocSeparators)
}

func (ye *yamlEncoder) PrintLeadingContent(writer io.Writer, content string) error {
	return PrintYAMLLeadingContent(writer, content, ye.prefs.PrintDocSeparators, ye.prefs.ColorsEnabled)
}

func (ye *yamlEncoder) Encode(writer io.Writer, node *CandidateNode) error {
	log.Debugf("encoderYaml - going to print %v", NodeToString(node))
	// Detect line ending style from LeadingContent
	lineEnding := "\n"
	if strings.Contains(node.LeadingContent, "\r\n") {
		lineEnding = "\r\n"
	}
	if node.Kind == ScalarNode && ye.prefs.UnwrapScalar {
		valueToPrint := node.Value
		if node.LeadingContent == "" || valueToPrint != "" {
			valueToPrint = valueToPrint + lineEnding
		}
		return writeString(writer, valueToPrint)
	}

	target, err := node.MarshalYAML()
	if err != nil {
		return err
	}

	trailingContent := target.FootComment
	target.FootComment = ""

	// go.yaml.in/yaml/v4's emitter decides whether a multi-line string may use
	// literal/folded block style by scanning its bytes for "printable" characters,
	// but that scan never recognises 4-byte UTF-8 sequences (runes above U+FFFF,
	// e.g. emoji outside the Basic Multilingual Plane) as printable. That makes it
	// treat such runes as special characters and silently fall back to a
	// double-quoted, single-line-with-escapes rendering. We work around this
	// upstream bug by swapping those runes out for placeholder runes the emitter
	// does recognise as printable before dumping, then swapping the original runes
	// back into the encoded bytes afterwards.
	unicodeWorkaround := newSupplementaryRuneWorkaround()
	unicodeWorkaround.remapNode(target)

	// Always buffered. Both post-encoding passes below — restoring the
	// remapped runes, and re-indenting a single-quoted scalar's closing
	// quote — work on the finished bytes, so the document cannot stream
	// straight to the writer regardless of colour or workaround state.
	destination := bytes.NewBuffer(nil)

	indent := ye.prefs.Indent
	if indent < 2 {
		indent = 2
	} else if indent > 9 {
		indent = 9
	}

	dumper, err := yaml.NewDumper(destination,
		yaml.WithV3Defaults(),
		yaml.WithIndent(indent),
		yaml.WithCompactSeqIndent(ye.prefs.CompactSequenceIndent),
		yaml.WithLineWidth(-1),
	)
	if err != nil {
		return fmt.Errorf("configure YAML encoding: %w", err)
	}

	err = dumper.Dump(target)
	if closeErr := dumper.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}

	if err := ye.PrintLeadingContent(destination, trailingContent); err != nil {
		return err
	}

	// Restore before re-indenting, not after: the indent fix measures
	// leading spaces on the bytes it is handed, and restoring the original
	// runes is what makes those bytes final. Reversing the order would
	// re-indent a document that still contained placeholders.
	encoded := fixSingleQuotedClosingIndent(unicodeWorkaround.restore(destination.Bytes()))

	if ye.prefs.ColorsEnabled {
		return colorizeAndPrint(encoded, writer)
	}
	return writeString(writer, string(encoded))
}

// fixSingleQuotedClosingIndent corrects the closing quote line of a block-style
// single-quoted scalar whose value ends with a blank line. go-yaml emits that
// closing quote flush against column 0 instead of lining it up with the rest of
// the scalar's continuation lines, which some YAML tools reject. This walks the
// already-encoded document line by line, tracking single-quoted scalars that
// span multiple lines, and re-indents an under-indented lone closing quote to
// match the indentation of the scalar's most recent non-blank continuation line.
func fixSingleQuotedClosingIndent(data []byte) []byte {
	lines := strings.Split(string(data), "\n")

	inSingleQuoted := false
	continuationIndent := 0

	for i, rawLine := range lines {
		line := rawLine
		hasCR := strings.HasSuffix(line, "\r")
		if hasCR {
			line = line[:len(line)-1]
		}

		// doubled '' is an escaped literal quote inside the scalar, not a delimiter
		quoteCount := strings.Count(strings.ReplaceAll(line, "''", ""), "'")
		closesHere := inSingleQuoted && quoteCount%2 == 1

		if inSingleQuoted {
			trimmed := strings.TrimSpace(line)
			switch {
			case closesHere && trimmed == "'":
				leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))
				if leadingSpaces < continuationIndent {
					newLine := strings.Repeat(" ", continuationIndent) + strings.TrimLeft(line, " ")
					if hasCR {
						newLine += "\r"
					}
					lines[i] = newLine
				}
			case !closesHere && trimmed != "":
				continuationIndent = len(line) - len(strings.TrimLeft(line, " "))
			}

			if closesHere {
				inSingleQuoted = false
				continuationIndent = 0
			}
		} else if quoteCount%2 == 1 {
			inSingleQuoted = true
			continuationIndent = 0
		}
	}

	return []byte(strings.Join(lines, "\n"))
}
