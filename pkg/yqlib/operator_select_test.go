package yqlib

import (
	"testing"
)

var selectOperatorScenarios = []expressionScenario{
	{
		skipDoc:    true,
		document:   `cat: pants`,
		expression: `select(.nope) | key + " why though?"`,
		expected:   []string{},
	},
	{
		skipDoc:    true,
		document:   `cat`,
		expression: `select(false, true)`,
		expected: []string{
			"D0, P[], (!!str)::cat\n",
		},
	},
	{
		skipDoc:    true,
		document:   `cat`,
		expression: `select(true, false)`,
		expected: []string{
			"D0, P[], (!!str)::cat\n",
		},
	},
	{
		skipDoc:    true,
		document:   `cat`,
		expression: `select(false)`,
		expected:   []string{},
	},
	{
		description: "Select elements from array using wildcard prefix",
		document:    `[cat,goat,dog]`,
		expression:  `.[] | select(. == "*at")`,
		expected: []string{
			"D0, P[0], (!!str)::cat\n",
			"D0, P[1], (!!str)::goat\n",
		},
	},
	{
		description: "Select elements from array using wildcard suffix",
		document:    `[go-kart,goat,dog]`,
		expression:  `.[] | select(. == "go*")`,
		expected: []string{
			"D0, P[0], (!!str)::go-kart\n",
			"D0, P[1], (!!str)::goat\n",
		},
	},
	{
		description: "Select elements from array using wildcard prefix and suffix",
		document:    `[ago, go, meow, going]`,
		expression:  `.[] | select(. == "*go*")`,
		expected: []string{
			"D0, P[0], (!!str)::ago\n",
			"D0, P[1], (!!str)::go\n",
			"D0, P[3], (!!str)::going\n",
		},
	},
	{
		description:    "Select elements from array with regular expression",
		subdescription: "See more regular expression examples under the [`string` operator docs](https://mikefarah.gitbook.io/yq/operators/string-operators).",
		document:       `[this_0, not_this, nor_0_this, thisTo_4]`,
		expression:     `.[] | select(test("[a-zA-Z]+_[0-9]$"))`,
		expected: []string{
			"D0, P[0], (!!str)::this_0\n",
			"D0, P[3], (!!str)::thisTo_4\n",
		},
	},
	{
		skipDoc:    true,
		document:   "a: hello",
		document2:  "b: world",
		expression: `select(.a == "hello" or .b == "world")`,
		expected: []string{
			"D0, P[], (!!map)::a: hello\n",
			"D0, P[], (!!map)::b: world\n",
		},
	},
	{
		description: "select splat",
		skipDoc:     true,
		document:    "a: hello",
		document2:   "b: world",
		expression:  `select(.a == "hello" or .b == "world")[]`,
		expected: []string{
			"D0, P[a], (!!str)::hello\n",
			"D0, P[b], (!!str)::world\n",
		},
	},
	{
		description: "select does not update the map",
		skipDoc:     true,
		document:    `[{animal: cat, legs: {cool: true}}, {animal: fish}]`,
		expression:  `(.[] | select(.legs.cool == true).canWalk) = true | (.[] | .alive.things) = "yes"`,
		expected: []string{
			"D0, P[], (!!seq)::[{animal: cat, legs: {cool: true}, canWalk: true, alive: {things: yes}}, {animal: fish, alive: {things: yes}}]\n",
		},
	},
	{
		skipDoc:    true,
		document:   `[hot, fot, dog]`,
		expression: `.[] | select(. == "*at")`,
		expected:   []string{},
	},
	{
		skipDoc:    true,
		document:   `a: [cat,goat,dog]`,
		expression: `.a.[] | select(. == "*at")`,
		expected: []string{
			"D0, P[a 0], (!!str)::cat\n",
			"D0, P[a 1], (!!str)::goat\n"},
	},
	{
		description: "Select items from a map",
		document:    `{ things: cat, bob: goat, horse: dog }`,
		expression:  `.[] | select(. == "cat" or test("og$"))`,
		expected: []string{
			"D0, P[things], (!!str)::cat\n",
			"D0, P[horse], (!!str)::dog\n",
		},
	},
	{
		description: "Use select and with_entries to filter map keys",
		document:    `{name: bob, legs: 2, game: poker}`,
		expression:  `with_entries(select(.key | test("ame$")))`,
		expected: []string{
			"D0, P[], (!!map)::name: bob\ngame: poker\n",
		},
	},
	{
		description:    "Select multiple items in a map and update",
		subdescription: "Note the brackets around the entire LHS.",
		document:       `a: { things: cat, bob: goat, horse: dog }`,
		expression:     `(.a.[] | select(. == "cat" or . == "goat")) |= "rabbit"`,
		expected: []string{
			"D0, P[], (!!map)::a: {things: rabbit, bob: rabbit, horse: dog}\n",
		},
	},
	{
		skipDoc:    true,
		document:   `a: { things: {include: true}, notMe: {include: false}, andMe: {include: fold} }`,
		expression: `.a.[] | select(.include)`,
		expected: []string{
			"D0, P[a things], (!!map)::{include: true}\n",
			"D0, P[a andMe], (!!map)::{include: fold}\n",
		},
	},
	{
		skipDoc:    true,
		document:   `[cat,~,dog]`,
		expression: `.[] | select(. == ~)`,
		expected: []string{
			"D0, P[1], (!!null)::~\n",
		},
	},
	{
		// Regression test: missing keys are dropped from [...] inside select.
		// select evaluates its condition against a read-only context
		// (SingleReadonlyChildContext, context.go), which suppresses the auto-created
		// null node in traverseMap (operator_traverse_path.go) for a missing key. So
		// [.a] is currently [] (length 0) here instead of [null] (length 1), as it is
		// at the top level, and this select produces no output instead of {}.
		skipDoc:     true,
		description: "select with a missing key collected into an array - currently broken",
		expression:  `{} | select([.a] | length == 1)`,
		expected: []string{
			"D0, P[], (!!map)::{}\n",
		},
	},
	{
		// Contrast case: indexing an array always fills in missing elements with null,
		// regardless of the read-only context, so this already works and should keep
		// working once the length regression above is fixed.
		skipDoc:     true,
		description: "select with a missing key collected then indexed - already correct",
		expression:  `{} | select((([.a] | .[0]) | tag) == "!!null")`,
		expected: []string{
			"D0, P[], (!!map)::{}\n",
		},
	},
}

func TestSelectOperatorScenarios(t *testing.T) {
	for _, tt := range selectOperatorScenarios {
		testScenario(t, &tt)
	}
	documentOperatorScenarios(t, "select", selectOperatorScenarios)
}
