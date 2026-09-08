package yqlib

import (
	"testing"

	"github.com/mikefarah/yq/v4/test"
)

var yamlFormatScenarios = []formatScenario{
	{
		description: "scalar with doc separator",
		skipDoc:     true,
		input:       "--- cat",
		expected:    "---\ncat\n",
	},
	{
		description: "CRLF doc separator",
		skipDoc:     true,
		input:       "---\r\ncat\r\n",
		expected:    "---\r\ncat\r\n",
	},
	{
		description: "yaml directive preserved (LF)",
		skipDoc:     true,
		input:       "%YAML 1.1\n---\ncat\n",
		expected:    "%YAML 1.1\n---\ncat\n",
	},
	{
		description: "yaml directive preserved (CRLF)",
		skipDoc:     true,
		input:       "%YAML 1.1\r\n---\r\ncat\r\n",
		expected:    "%YAML 1.1\r\n---\r\ncat\r\n",
	},
	{
		description: "comment only no trailing newline",
		skipDoc:     true,
		input:       "# hello",
		expected:    "# hello\n",
	},

	{
		description: "scalar with doc separator",
		skipDoc:     true,
		input:       "---cat",
		expected:    "---cat\n",
	},
	{
		description: "basic - null",
		skipDoc:     true,
		input:       "null",
		expected:    "null\n",
	},
	{
		description: "basic - ~",
		skipDoc:     true,
		input:       "~",
		expected:    "~\n",
	},
	{
		description: "octal",
		skipDoc:     true,
		input:       "0o30",
		expression:  "tag",
		expected:    "!!int\n",
	},
	{
		description: "basic - [null]",
		skipDoc:     true,
		input:       "[null]",
		expected:    "[null]\n",
	},
	{
		description: "multi document anchor map",
		skipDoc:     true,
		input:       "a: &remember mike\n---\nb: *remember",
		expression:  "explode(.)",
		expected:    "a: mike\n---\nb: mike\n",
	},
	{
		description: "basic - [~]",
		skipDoc:     true,
		input:       "[~]",
		expected:    "[~]\n",
	},
	{
		description: "basic - null map value",
		skipDoc:     true,
		input:       "a: null",
		expected:    "a: null\n",
	},
	{
		description: "basic - number",
		skipDoc:     true,
		input:       "3",
		expected:    "3\n",
	},
	{
		description: "basic - float",
		skipDoc:     true,
		input:       "3.1",
		expected:    "3.1\n",
	},
	{
		description: "basic - float",
		skipDoc:     true,
		input:       "[1, 2]",
		expected:    "[1, 2]\n",
	},
	{
		description: "long line",
		skipDoc:     true,
		input:       "field: Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt",
		expected:    "field: Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt\n",
	},
	{
		description: "single quoted multi-line scalar ending with a blank line keeps closing quote indented",
		skipDoc:     true,
		input:       "paths:\n  /v1/quack:\n    get:\n      description: '\n        Get quacked.\n\n        - Duck\n\n        '\n",
		expected:    "paths:\n  /v1/quack:\n    get:\n      description: ' Get quacked.\n\n        - Duck\n\n        '\n",
	},
	{
		description: "single quoted multi-line scalar not ending with a blank line is unchanged",
		skipDoc:     true,
		input:       "description: '\n  Get quacked.\n\n  - Duck'\n",
		expected:    "description: ' Get quacked.\n\n  - Duck'\n",
	},
	{
		description: "single quoted multi-line scalar with a doubled quote escape is unchanged",
		skipDoc:     true,
		input:       "description: '\n  Get quacked, it''s fine.\n\n  - Duck'\n",
		expected:    "description: ' Get quacked, it''s fine.\n\n  - Duck'\n",
	},
	{
		description: "multiline string with a supplementary-plane emoji stays a literal block",
		skipDoc:     true,
		input:       "longDescription: |\n  Multiline file containing\n\n  🚧\n\n  A \"weird\" utf8 char?\n",
		expected:    "longDescription: |\n  Multiline file containing\n\n  🚧\n\n  A \"weird\" utf8 char?\n",
	},
	{
		description: "multiline string with a BMP symbol stays a literal block",
		skipDoc:     true,
		input:       "longDescription: |\n  Multiline file containing\n\n  ✅\n\n  A \"less weird\" utf8 char?\n",
		expected:    "longDescription: |\n  Multiline file containing\n\n  ✅\n\n  A \"less weird\" utf8 char?\n",
	},
}

var yamlParseScenarios = []expressionScenario{
	// {
	// 	description: "with a unquoted question mark in the string",
	// 	document:    "foo: {bar: a?bc}",
	// 	expected: []string{
	// 		"D0, P[], (!!map)::a: hello # things\n",
	// 	},
	// },
	{
		document: `a: hello # things`,
		expected: []string{
			"D0, P[], (!!map)::a: hello # things\n",
		},
	},
	{
		document:   "a: &a apple\nb: *a",
		expression: ".b | explode(.)",
		expected: []string{
			"D0, P[b], (!!str)::apple\n",
		},
	},
	{
		document: `a: [1,2]`,
		expected: []string{
			"D0, P[], (!!map)::a: [1, 2]\n",
		},
	},
	{
		document: `a: !horse [a]`,
		expected: []string{
			"D0, P[], (!!map)::a: !horse [a]\n",
		},
	},
	{
		description: "duplicate keys, last value wins",
		document:    `{"foo":1,"foo":2}`,
		expected: []string{
			"D0, P[], (!!map)::{\"foo\": 2}\n",
		},
	},
	{
		description: "duplicate keys, last value wins, pretty printed",
		document:    `{"foo":1,"foo":2}`,
		expression:  `... style=""`,
		expected: []string{
			"D0, P[], (!!map)::foo: 2\n",
		},
	},
}

func testYamlScenario(t *testing.T, s formatScenario) {
	test.AssertResultWithContext(t, s.expected, mustProcessFormatScenario(s, NewYamlDecoder(ConfiguredYamlPreferences), NewYamlEncoder(ConfiguredYamlPreferences)), s.description)
}

func TestYamlParseScenarios(t *testing.T) {
	for _, tt := range yamlParseScenarios {
		testScenario(t, &tt)
	}
}

func TestYamlFormatScenarios(t *testing.T) {
	for _, tt := range yamlFormatScenarios {
		testYamlScenario(t, tt)
	}
}
