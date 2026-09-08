package yqlib

import (
	"testing"
)

var styleOperatorScenarios = []expressionScenario{
	{
		description: "Update and set style of a particular node (simple)",
		document:    `a: {b: thing, c: something}`,
		expression:  `.a.b = "new" | .a.b style="double"`,
		expected: []string{
			"D0, P[], (!!map)::a: {b: \"new\", c: something}\n",
		},
	},
	{
		description: "Update and set style of a particular node using path variables",
		document:    `a: {b: thing, c: something}`,
		expression:  `with(.a.b ; . = "new" | . style="double")`,
		expected: []string{
			"D0, P[], (!!map)::a: {b: \"new\", c: something}\n",
		},
	},
	{
		description: "Set tagged style",
		document:    `{a: cat, b: 5, c: 3.2, e: true, f: [1,2,3], g: { something: cool}}`,
		expression:  `.. style="tagged"`,
		expected: []string{
			"D0, P[], (!!map)::!!map\na: !!str cat\nb: !!int 5\nc: !!float 3.2\ne: !!bool true\nf: !!seq\n    - !!int 1\n    - !!int 2\n    - !!int 3\ng: !!map\n    something: !!str cool\n",
		},
	},
	{
		description: "Set double quote style",
		document:    `{a: cat, b: 5, c: 3.2, e: true, f: [1,2,3], g: { something: cool}}`,
		expression:  `.. style="double"`,
		expected: []string{
			"D0, P[], (!!map)::a: \"cat\"\nb: \"5\"\nc: \"3.2\"\ne: \"true\"\nf:\n    - \"1\"\n    - \"2\"\n    - \"3\"\ng:\n    something: \"cool\"\n",
		},
	},
	{
		description: "Set double quote style on map keys too",
		document:    `{a: cat, b: 5, c: 3.2, e: true, f: [1,2,3], g: { something: cool}}`,
		expression:  `... style="double"`,
		expected: []string{
			"D0, P[], (!!map)::\"a\": \"cat\"\n\"b\": \"5\"\n\"c\": \"3.2\"\n\"e\": \"true\"\n\"f\":\n    - \"1\"\n    - \"2\"\n    - \"3\"\n\"g\":\n    \"something\": \"cool\"\n",
		},
	},
	{
		skipDoc:    true,
		document:   "bing: &foo {x: z}\na:\n  c: cat\n  <<: [*foo]",
		expression: `(... | select(tag=="!!str")) style="single"`,
		expected: []string{
			"D0, P[], (!!map)::'bing': &foo {'x': 'z'}\n'a':\n    'c': 'cat'\n    <<: [*foo]\n",
		},
	},
	{
		description: "Set single quote style",
		document:    `{a: cat, b: 5, c: 3.2, e: true, f: [1,2,3], g: { something: cool}}`,
		expression:  `.. style="single"`,
		expected: []string{
			"D0, P[], (!!map)::a: 'cat'\nb: '5'\nc: '3.2'\ne: 'true'\nf:\n    - '1'\n    - '2'\n    - '3'\ng:\n    something: 'cool'\n",
		},
	},
	{
		description: "Set literal quote style",
		document:    `{a: cat, b: 5, c: 3.2, e: true, f: [1,2,3], g: { something: cool}}`,
		expression:  `.. style="literal"`,
		expected: []string{
			`D0, P[], (!!map)::a: |-
    cat
b: |-
    5
c: |-
    3.2
e: |-
    true
f:
    - |-
      1
    - |-
      2
    - |-
      3
g:
    something: |-
        cool
`,
		},
	},
	{
		description: "Set literal quote style, preserving trailing whitespace before a newline",
		document:    `a: "good \"bye \n    cruel\" world!"`,
		expression:  `.a style="literal"`,
		expected: []string{
			"D0, P[], (!!map)::a: |-\n    good \"bye \n        cruel\" world!\n",
		},
	},
	{
		description: "Set folded quote style",
		document:    `{a: cat, b: 5, c: 3.2, e: true, f: [1,2,3], g: { something: cool}}`,
		expression:  `.. style="folded"`,
		expected: []string{
			`D0, P[], (!!map)::a: >-
    cat
b: >-
    5
c: >-
    3.2
e: >-
    true
f:
    - >-
      1
    - >-
      2
    - >-
      3
g:
    something: >-
        cool
`,
		},
	},
	{
		description: "Set flow quote style",
		document:    `{a: cat, b: 5, c: 3.2, e: true, f: [1,2,3], g: { something: cool}}`,
		expression:  `.. style="flow"`,
		expected: []string{
			"D0, P[], (!!map)::{a: cat, b: 5, c: 3.2, e: true, f: [1, 2, 3], g: {something: cool}}\n",
		},
	},
	{
		description:           "Reset style - or pretty print",
		subdescription:        "Set empty (default) quote style, note the usage of `...` to match keys too. Note that there is a `--prettyPrint/-P` short flag for this. Note that double quoted strings that need to escape characters (e.g. a newline) will keep their double quotes.",
		dontFormatInputForDoc: true,
		document:              `{a: cat, "b": 5, 'c': 3.2, "e": true,  f: [1,2,3], "g": { something: "cool"}, "h": "double\nquote" }`,
		expression:            `... style=""`,
		expected: []string{
			"D0, P[], (!!map)::a: cat\nb: 5\nc: 3.2\ne: true\nf:\n    - 1\n    - 2\n    - 3\ng:\n    something: cool\nh: \"double\\nquote\"\n",
		},
	},
	{
		skipDoc:     true,
		description: "Reset style keeps double quotes on keys that need escaping",
		document:    "\"double\\nquote\": true\n",
		expression:  `... style=""`,
		expected: []string{
			"D0, P[], (!!map)::? \"double\\nquote\"\n: true\n",
		},
	},
	{
		skipDoc:     true,
		description: "Reset style unquotes plain and single quoted values without special characters",
		document:    "a: plain\nb: 'single'\n",
		expression:  `... style=""`,
		expected: []string{
			"D0, P[], (!!map)::a: plain\nb: single\n",
		},
	},
	{
		skipDoc:     true,
		description: "Reset style still quotes values that need it for other reasons",
		document:    "a: \"{needs quote}\"\n",
		expression:  `... style=""`,
		expected: []string{
			"D0, P[], (!!map)::a: '{needs quote}'\n",
		},
	},
	{
		skipDoc:     true,
		description: "Reset style leaves existing literal block scalars alone",
		document:    "a: |-\n  cat\n  dog\n",
		expression:  `... style=""`,
		expected: []string{
			"D0, P[], (!!map)::a: |-\n    cat\n    dog\n",
		},
	},
	{
		description: "Set style relatively with assign-update",
		document:    `{a: single, b: double}`,
		expression:  `.[] style |= .`,
		expected: []string{
			"D0, P[], (!!map)::{a: 'single', b: \"double\"}\n",
		},
	},
	{
		skipDoc:    true,
		document:   `{a: cat, b: double}`,
		expression: `.a style=.b`,
		expected: []string{
			"D0, P[], (!!map)::{a: \"cat\", b: double}\n",
		},
	},
	{
		description:           "Read style",
		document:              `{a: "cat", b: 'thing'}`,
		dontFormatInputForDoc: true,
		expression:            `.. | style`,
		expected: []string{
			"D0, P[], (!!str)::flow\n",
			"D0, P[a], (!!str)::double\n",
			"D0, P[b], (!!str)::single\n",
		},
	},
	{
		skipDoc:    true,
		document:   `a: cat`,
		expression: `.. | style`,
		expected: []string{
			"D0, P[], (!!str)::\n",
			"D0, P[a], (!!str)::\n",
		},
	},
}

func TestStyleOperatorScenarios(t *testing.T) {
	for _, tt := range styleOperatorScenarios {
		testScenario(t, &tt)
	}
	documentOperatorScenarios(t, "style", styleOperatorScenarios)
}
