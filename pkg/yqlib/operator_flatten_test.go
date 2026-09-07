package yqlib

import (
	"testing"
)

var flattenOperatorScenarios = []expressionScenario{
	{
		description:    "Flatten",
		subdescription: "Recursively flattens all arrays",
		document:       `[1, [2], [[3]]]`,
		expression:     `flatten`,
		expected: []string{
			"D0, P[], (!!seq)::[1, 2, 3]\n",
		},
	},
	{
		description: "Flatten splat",
		skipDoc:     true,
		document:    `[1, [2], [[3]]]`,
		expression:  `flatten[]`,
		expected: []string{
			"D0, P[0], (!!int)::1\n",
			"D0, P[1], (!!int)::2\n",
			"D0, P[2], (!!int)::3\n",
		},
	},
	{
		description: "Flatten with depth of one",
		document:    `[1, [2], [[3]]]`,
		expression:  `flatten(1)`,
		expected: []string{
			"D0, P[], (!!seq)::[1, 2, [3]]\n",
		},
	},
	{
		description: "Flatten with depth and splat",
		skipDoc:     true,
		document:    `[1, [2], [[3]]]`,
		expression:  `flatten(1)[]`,
		expected: []string{
			"D0, P[0], (!!int)::1\n",
			"D0, P[1], (!!int)::2\n",
			"D0, P[2], (!!seq)::[3]\n",
		},
	},
	{
		description: "Flatten empty array",
		document:    `[[]]`,
		expression:  `flatten`,
		expected: []string{
			"D0, P[], (!!seq)::[]\n",
		},
	},
	{
		description: "Flatten array of objects",
		document:    `[{foo: bar}, [{foo: baz}]]`,
		expression:  `flatten`,
		expected: []string{
			"D0, P[], (!!seq)::[{foo: bar}, {foo: baz}]\n",
		},
	},
	{
		description:    "Flatten resolves an alias to a sequence",
		subdescription: "Aliased sequences are dereferenced and inlined, just like an inline sequence.",
		document:       `{base_list: &base [item1, item2], extended_list: [*base, item3]}`,
		expression:     `.extended_list | flatten`,
		expected: []string{
			"D0, P[extended_list], (!!seq)::[item1, item2, item3]\n",
		},
	},
	{
		description: "Flatten resolves a chain of aliased sequences",
		skipDoc:     true,
		document:    `[&x [1, 2], &y [*x, 3], *y, 4]`,
		expression:  `flatten`,
		expected: []string{
			"D0, P[], (!!seq)::[1, 2, 1, 2, 3, 1, 2, 3, 4]\n",
		},
	},
	{
		description: "Flatten with depth limit still resolves alias children",
		skipDoc:     true,
		document:    `[&x [1, [2, 3]], *x, 4]`,
		expression:  `flatten(1)`,
		expected: []string{
			"D0, P[], (!!seq)::[1, [2, 3], 1, [2, 3], 4]\n",
		},
	},
	{
		description:    "Flatten leaves aliases to non-sequences untouched",
		subdescription: "Only aliases that resolve to sequences are dereferenced; other aliases are left as-is.",
		document:       `{m: &m {a: 1}, s: &s cat, l: [*m, *s, 2]}`,
		expression:     `.l | flatten`,
		expected: []string{
			"D0, P[l], (!!seq)::[*m, *s, 2]\n",
		},
	},
}

func TestFlattenOperatorScenarios(t *testing.T) {
	for _, tt := range flattenOperatorScenarios {
		testScenario(t, &tt)
	}
	documentOperatorScenarios(t, "flatten", flattenOperatorScenarios)
}
