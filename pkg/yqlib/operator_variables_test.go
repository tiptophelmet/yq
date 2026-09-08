package yqlib

import (
	"testing"
)

var variableOperatorScenarios = []expressionScenario{
	{
		skipDoc:    true,
		document:   `{}`,
		expression: `.a.b as $foo | .`,
		expected: []string{
			"D0, P[], (!!map)::{}\n",
		},
	},
	{
		skipDoc:       true,
		document:      `{}`,
		expression:    `.a.b as $foo`,
		expectedError: "must use variable with a pipe, e.g. `exp as $x | ...`",
	},
	{
		document:      "a: [cat]",
		skipDoc:       true,
		expression:    "(.[] | {.name: .}) as $item | .",
		expectedError: `cannot index array with 'name' (strconv.ParseInt: parsing "name": invalid syntax)`,
	},
	{
		description: "Single value variable",
		document:    `a: cat`,
		expression:  `.a as $foo | $foo`,
		expected: []string{
			"D0, P[a], (!!str)::cat\n",
		},
	},
	{
		description: "Multi value variable",
		document:    `[cat, dog]`,
		expression:  `.[] as $foo | $foo`,
		expected: []string{
			"D0, P[0], (!!str)::cat\n",
			"D0, P[1], (!!str)::dog\n",
		},
	},
	{
		skipDoc:    true,
		document:   `[1, 2]`,
		expression: `.[] | . as $f | select($f == 2)`,
		expected: []string{
			"D0, P[1], (!!int)::2\n",
		},
	},
	{
		skipDoc:    true,
		document:   `[1, 2]`,
		expression: `[.[] | . as $f | $f + 1]`,
		expected: []string{
			"D0, P[], (!!seq)::- 2\n- 3\n",
		},
	},
	{
		description:    "Using variables as a lookup",
		subdescription: "Example taken from [jq](https://stedolan.github.io/jq/manual/#Variable/SymbolicBindingOperator:...as$identifier|...)",
		document: `{"posts": [{"title": "First post", "author": "anon"},
			{"title": "A well-written article", "author": "person1"}],
	"realnames": {"anon": "Anonymous Coward",
					"person1": "Person McPherson"}}`,
		expression: `.realnames as $names | .posts[] | {"title":.title, "author": $names[.author]}`,
		expected: []string{
			"D0, P[], (!!map)::title: \"First post\"\nauthor: \"Anonymous Coward\"\n",
			"D0, P[], (!!map)::title: \"A well-written article\"\nauthor: \"Person McPherson\"\n",
		},
	},
	{
		description: "Using variables to swap values",
		document:    "a: a_value\nb: b_value",
		expression:  `.a as $x  | .b as $y | .b = $x | .a = $y`,
		expected: []string{
			"D0, P[], (!!map)::a: b_value\nb: a_value\n",
		},
	},
	{
		description:    "Use ref to reference a path repeatedly",
		subdescription: "Note: You may find the `with` operator more useful.",
		document:       `a: {b: thing, c: something}`,
		expression:     `.a.b ref $x | $x = "new" | $x style="double"`,
		expected: []string{
			"D0, P[], (!!map)::a: {b: \"new\", c: something}\n",
		},
	},
	{
		skipDoc:     true,
		description: "Look up every value using keys, for a sequence",
		document:    `["a","b"]`,
		expression:  `. as $o | keys[] | $o[.]`,
		expected: []string{
			"D0, P[0], (!!str)::a\n",
			"D0, P[1], (!!str)::b\n",
		},
	},
	{
		skipDoc:     true,
		description: "Look up every value using keys, for a map",
		document:    `{a: 1, b: 2}`,
		expression:  `. as $o | keys[] | $o[.]`,
		expected: []string{
			"D0, P[a], (!!int)::1\n",
			"D0, P[b], (!!int)::2\n",
		},
	},
	{
		skipDoc:     true,
		description: "Look up multiple indices per candidate, for every candidate",
		document:    `["a","b"]`,
		expression:  `. as $o | keys[] | $o[., 0]`,
		expected: []string{
			"D0, P[0], (!!str)::a\n",
			"D0, P[1], (!!str)::b\n",
		},
	},
	{
		skipDoc:     true,
		description: "Nested ref bindings ending in an assignment do not multiply the document",
		document:    `top: [{n: d1, f: f1, h: [{hn: h1a}, {hn: h1b}]}, {n: d2, f: f2, h: [{hn: h2a}, {hn: h2b}]}]`,
		expression:  `.top[] ref $p | $p.h[] ref $h | $h.hf = $p.f`,
		expected: []string{
			"D0, P[], (!!map)::top: [{n: d1, f: f1, h: [{hn: h1a, hf: f1}, {hn: h1b, hf: f1}]}, {n: d2, f: f2, h: [{hn: h2a, hf: f2}, {hn: h2b, hf: f2}]}]\n",
		},
	},
	{
		skipDoc:     true,
		description: "Nested ref binding with an in-place update still yields a single result",
		document:    `top: [{n: d1, f: f1, h: [{hn: h1a}, {hn: h1b}]}, {n: d2, f: f2, h: [{hn: h2a}, {hn: h2b}]}]`,
		expression:  `.top[] ref $t | $t |= .h[] |= (.f = $t.f)`,
		expected: []string{
			"D0, P[], (!!map)::top: [{n: d1, f: f1, h: [{hn: h1a, f: f1}, {hn: h1b, f: f1}]}, {n: d2, f: f2, h: [{hn: h2a, f: f2}, {hn: h2b, f: f2}]}]\n",
		},
	},
	{
		skipDoc:     true,
		description: "Ref bindings for genuinely distinct matches are not de-duplicated",
		document:    `top: [{n: d1}, {n: d2}]`,
		expression:  `.top[] ref $p | $p.n`,
		expected: []string{
			"D0, P[top 0 n], (!!str)::d1\n",
			"D0, P[top 1 n], (!!str)::d2\n",
		},
	},
}

func TestVariableOperatorScenarios(t *testing.T) {
	for _, tt := range variableOperatorScenarios {
		testScenario(t, &tt)
	}
	documentOperatorScenarios(t, "variable-operators", variableOperatorScenarios)
}
