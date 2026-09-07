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

	destination := writer
	tempBuffer := bytes.NewBuffer(nil)
	if ye.prefs.ColorsEnabled || unicodeWorkaround.active() {
		destination = tempBuffer
	}

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

	if ye.prefs.ColorsEnabled {
		return colorizeAndPrint(unicodeWorkaround.restore(tempBuffer.Bytes()), writer)
	}
	if unicodeWorkaround.active() {
		return writeString(writer, string(unicodeWorkaround.restore(tempBuffer.Bytes())))
	}
	return nil
}
