package yqlib

import (
	"regexp"
	"strings"
)

// foldedHeaderRe matches the tail of a line that opens a folded (">") block
// scalar: a '>' preceded by whitespace, ':' or '-' (or at the very start of
// the trimmed line), followed by an optional chomping indicator ('-' or '+')
// and/or explicit indentation indicator, in either order.
var foldedHeaderRe = regexp.MustCompile(`(?:^|[\s:-])>([0-9][+-]?|[+-][0-9]?|[+-]?)$`)

// fixFoldedBlockBlankLines works around a bug in go.yaml.in/yaml/v4's folded
// scalar writer: it emits an extra line break for every newline in the
// scalar whenever the preceding written line did not start with whitespace.
// When the following content line is more-indented than the block's base
// indentation, that newline is already preserved by the extra indentation,
// so the emitter's extra break surfaces as a phantom blank line - and the
// same thing happens once more at the very end of the block. This walks the
// already-encoded document line by line, finds folded block scalars, and
// drops those emitter-inserted blank lines while leaving genuine blank
// lines (which folded style encodes as two consecutive breaks) untouched.
func fixFoldedBlockBlankLines(data []byte) []byte {
	// Split(..., "\n") always yields a trailing "" element whenever the data
	// itself ends in "\n" - that sentinel isn't a line, real or blank, and
	// must not be handed to the end-of-block cleanup below or it gets
	// mistaken for (and dropped as) an emitter-inserted trailing blank line.
	text := string(data)
	trailingNewline := strings.HasSuffix(text, "\n")
	if trailingNewline {
		text = text[:len(text)-1]
	}

	lines := strings.Split(text, "\n")
	var out []string

	for i := 0; i < len(lines); {
		line := lines[i]
		if !isFoldedBlockHeader(line) {
			out = append(out, line)
			i++
			continue
		}

		headerIndent := leadingSpaceCount(line)
		out = append(out, line)
		i++

		start := i
		for i < len(lines) {
			if strings.TrimSpace(lines[i]) != "" && leadingSpaceCount(lines[i]) <= headerIndent {
				break
			}
			i++
		}
		out = append(out, cleanFoldedBlock(lines[start:i])...)
	}

	result := strings.Join(out, "\n")
	if trailingNewline {
		result += "\n"
	}
	return []byte(result)
}

// cleanFoldedBlock removes emitter-inserted blank lines from the content
// lines of a single folded block scalar (excluding its header line).
func cleanFoldedBlock(block []string) []string {
	baseIndent := -1
	for _, l := range block {
		if strings.TrimSpace(l) != "" {
			baseIndent = leadingSpaceCount(l)
			break
		}
	}
	if baseIndent < 0 {
		return block
	}

	var result []string
	for i := 0; i < len(block); {
		if strings.TrimSpace(block[i]) != "" {
			result = append(result, block[i])
			i++
			continue
		}

		runStart := i
		for i < len(block) && strings.TrimSpace(block[i]) == "" {
			i++
		}
		run := block[runStart:i]

		// Drop exactly one emitter-inserted blank line: either the next
		// content line is more-indented than the block's base (the newline
		// it represents is already preserved by that extra indentation), or
		// this run sits at the very end of the block (same bug, no "next"
		// line to be more-indented than).
		atEndOfBlock := i == len(block)
		nextIsDeeper := !atEndOfBlock && leadingSpaceCount(block[i]) > baseIndent
		if (atEndOfBlock || nextIsDeeper) && len(run) > 0 {
			run = run[1:]
		}
		result = append(result, run...)
	}
	return result
}

func isFoldedBlockHeader(line string) bool {
	trimmed := strings.TrimRight(line, " \t\r")
	if idx := strings.Index(trimmed, " #"); idx >= 0 {
		trimmed = strings.TrimRight(trimmed[:idx], " \t")
	}
	return foldedHeaderRe.MatchString(trimmed)
}

func leadingSpaceCount(s string) int {
	return len(s) - len(strings.TrimLeft(s, " "))
}
