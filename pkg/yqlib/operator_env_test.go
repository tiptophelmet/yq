package yqlib

import (
	"testing"
)

func TestEnvOperatorDynamicNameScenarios(t *testing.T) {
	t.Run("with(...) using key as the dynamic env var name", func(t *testing.T) {
		t.Setenv("FOO", "bar")
		t.Setenv("BAR", "baz")
		testScenario(t, &expressionScenario{
			description:    "Use each map key to look up its own environment variable",
			subdescription: "`env(...)` and `strenv(...)` can take any yq expression, not just a literal variable name. Here `key` picks the variable name for each entry.",
			skipDoc:        true,
			document:       `configmap: {values: {FOO: "", BAR: ""}}`,
			expression:     `with(.configmap.values[]; . = env(key))`,
			expected: []string{
				"D0, P[], (!!map)::configmap: {values: {FOO: \"bar\", BAR: \"baz\"}}\n",
			},
		})
	})

	t.Run("literal env(FOO) unaffected by dynamic support", func(t *testing.T) {
		t.Setenv("FOO", "bar")
		testScenario(t, &expressionScenario{
			description: "A literal, set variable name is still read exactly as before",
			skipDoc:     true,
			document:    `configmap: {values: {FOO: "", BAR: ""}}`,
			expression:  `with(.configmap.values[]; . = env(FOO))`,
			expected: []string{
				"D0, P[], (!!map)::configmap: {values: {FOO: \"bar\", BAR: \"bar\"}}\n",
			},
		})
	})

	t.Run("literal strenv(FOO) unaffected by dynamic support", func(t *testing.T) {
		t.Setenv("FOO", "true")
		testScenario(t, &expressionScenario{
			description: "A literal, set variable name is still read as a plain string",
			skipDoc:     true,
			expression:  `strenv(FOO)`,
			expected: []string{
				"D0, P[], (!!str)::true\n",
			},
		})
	})

	t.Run("strenv with dynamic name returns empty string when unset", func(t *testing.T) {
		testScenario(t, &expressionScenario{
			description: "strenv() with a dynamic name resolves to an empty string when the target variable isn't set",
			skipDoc:     true,
			document:    `a: unsetDynamicEnvVar`,
			expression:  `strenv(.a)`,
			expected: []string{
				"D0, P[], (!!str)::\n",
			},
		})
	})

	t.Run("env() dynamic name that cannot be parsed", func(t *testing.T) {
		testScenario(t, &expressionScenario{
			description:   "env() with an expression that fails to parse",
			skipDoc:       true,
			expression:    `env(&&&)`,
			expectedError: `could not process substitution '&&&' in env(): 1:1: lexer: invalid input text "&&&"`,
		})
	})

	t.Run("env() dynamic name resolving to multiple nodes", func(t *testing.T) {
		testScenario(t, &expressionScenario{
			description:   "env() dynamic expression must resolve to a single value",
			skipDoc:       true,
			document:      `a: [x, y]`,
			expression:    `env(.a[])`,
			expectedError: `substitution '.a[]' in env() must resolve to exactly one value, found 2`,
		})
	})
}

var envOperatorScenarios = []expressionScenario{
	{
		description:          "Read string environment variable",
		skipDoc:              true,
		environmentVariables: map[string]string{"myenv": "[cat,dog]"},
		expression:           `env(myenv)[]`,
		expected: []string{
			"D0, P[0], (!!str)::cat\n",
			"D0, P[1], (!!str)::dog\n",
		},
	},
	{
		description:          "Read string environment variable",
		environmentVariables: map[string]string{"myenv": "cat meow"},
		expression:           `.a = env(myenv)`,
		expected: []string{
			"D0, P[], ()::a: cat meow\n",
		},
	},
	{
		description:          "Read boolean environment variable",
		environmentVariables: map[string]string{"myenv": "true"},
		expression:           `.a = env(myenv)`,
		expected: []string{
			"D0, P[], ()::a: true\n",
		},
	},
	{
		description:          "Read numeric environment variable",
		environmentVariables: map[string]string{"myenv": "12"},
		expression:           `.a = env(myenv)`,
		expected: []string{
			"D0, P[], ()::a: 12\n",
		},
	},
	{
		description:          "Read yaml environment variable",
		environmentVariables: map[string]string{"myenv": "{b: fish}"},
		expression:           `.a = env(myenv)`,
		expected: []string{
			"D0, P[], ()::a: {b: fish}\n",
		},
	},
	{
		description:          "Read boolean environment variable as a string",
		environmentVariables: map[string]string{"myenv": "true"},
		expression:           `.a = strenv(myenv)`,
		expected: []string{
			"D0, P[], ()::a: \"true\"\n",
		},
	},
	{
		description:          "Read numeric environment variable as a string",
		environmentVariables: map[string]string{"myenv": "12"},
		expression:           `.a = strenv(myenv)`,
		expected: []string{
			"D0, P[], ()::a: \"12\"\n",
		},
	},
	{
		description:          "Dynamically update a path from an environment variable",
		subdescription:       "The env variable can be any valid yq expression.",
		document:             `{a: {b: [{name: dog}, {name: cat}]}}`,
		environmentVariables: map[string]string{"pathEnv": ".a.b[0].name", "valueEnv": "moo"},
		expression:           `eval(strenv(pathEnv)) = strenv(valueEnv)`,
		expected: []string{
			"D0, P[], (!!map)::{a: {b: [{name: moo}, {name: cat}]}}\n",
		},
	},
	{
		description:          "Dynamic key lookup with environment variable",
		environmentVariables: map[string]string{"myenv": "cat"},
		document:             `{cat: meow, dog: woof}`,
		expression:           `.[env(myenv)]`,
		expected: []string{
			"D0, P[cat], (!!str)::meow\n",
		},
	},
	{
		description:          "Replace strings with envsubst",
		environmentVariables: map[string]string{"myenv": "cat"},
		expression:           `"the ${myenv} meows" | envsubst`,
		expected: []string{
			"D0, P[], (!!str)::the cat meows\n",
		},
	},
	{
		description: "Replace strings with envsubst, missing variables",
		expression:  `"the ${myenvnonexisting} meows" | envsubst`,
		expected: []string{
			"D0, P[], (!!str)::the  meows\n",
		},
	},
	{
		description:    "Replace strings with envsubst(nu), missing variables",
		subdescription: "(nu) not unset, will fail if there are unset (missing) variables",
		expression:     `"the ${myenvnonexisting} meows" | envsubst(nu)`,
		expectedError:  "variable ${myenvnonexisting} not set",
	},
	{
		description:    "Replace strings with envsubst(ne), missing variables",
		subdescription: "(ne) not empty, only validates set variables",
		expression:     `"the ${myenvnonexisting} meows" | envsubst(ne)`,
		expected: []string{
			"D0, P[], (!!str)::the  meows\n",
		},
	},
	{
		description:          "Replace strings with envsubst(ne), empty variable",
		subdescription:       "(ne) not empty, will fail if a references variable is empty",
		environmentVariables: map[string]string{"myenv": ""},
		expression:           `"the ${myenv} meows" | envsubst(ne)`,
		expectedError:        "variable ${myenv} set but empty",
	},
	{
		description: "Replace strings with envsubst, missing variables with defaults",
		expression:  `"the ${myenvnonexisting-dog} meows" | envsubst`,
		expected: []string{
			"D0, P[], (!!str)::the dog meows\n",
		},
	},
	{
		description:    "Replace strings with envsubst(nu), missing variables with defaults",
		subdescription: "Having a default specified skips over the missing variable.",
		expression:     `"the ${myenvnonexisting-dog} meows" | envsubst(nu)`,
		expected: []string{
			"D0, P[], (!!str)::the dog meows\n",
		},
	},
	{
		description:          "Replace strings with envsubst(ne), missing variables with defaults",
		subdescription:       "Fails, because the variable is explicitly set to blank.",
		environmentVariables: map[string]string{"myEmptyEnv": ""},
		expression:           `"the ${myEmptyEnv-dog} meows" | envsubst(ne)`,
		expectedError:        "variable ${myEmptyEnv} set but empty",
	},
	{
		description:          "Replace string environment variable in document",
		environmentVariables: map[string]string{"myenv": "cat meow"},
		document:             "{v: \"${myenv}\"}",
		expression:           `.v |= envsubst`,
		expected: []string{
			"D0, P[], (!!map)::{v: \"cat meow\"}\n",
		},
	},
	{
		description:    "(Default) Return all envsubst errors",
		subdescription: "By default, all errors are returned at once.",
		expression:     `"the ${notThere} ${alsoNotThere}" | envsubst(nu)`,
		expectedError:  "variable ${notThere} not set\nvariable ${alsoNotThere} not set",
	},
	{
		description:   "Fail fast, return the first envsubst error (and abort)",
		expression:    `"the ${notThere} ${alsoNotThere}" | envsubst(nu,ff)`,
		expectedError: "variable ${notThere} not set",
	},
	{
		description:          "with header/footer",
		skipDoc:              true,
		environmentVariables: map[string]string{"myenv": "cat meow"},
		document:             "# abc\n{v: \"${myenv}\"}\n# xyz\n",
		expression:           `(.. | select(tag == "!!str")) |= envsubst`,
		expected: []string{
			"D0, P[], (!!map)::# abc\n{v: \"cat meow\"}\n# xyz\n",
		},
	},
}

func TestEnvOperatorScenarios(t *testing.T) {
	for _, tt := range envOperatorScenarios {
		testScenario(t, &tt)
	}
	documentOperatorScenarios(t, "env-variable-operators", envOperatorScenarios)
}

var envOperatorSecurityDisabledScenarios = []expressionScenario{
	{
		description:    "env() operation fails when security is enabled",
		subdescription: "Use `--security-disable-env-ops` to disable env operations for security.",
		expression:     `env("MYENV")`,
		expectedError:  "env operations have been disabled",
	},
	{
		description:    "strenv() operation fails when security is enabled",
		subdescription: "Use `--security-disable-env-ops` to disable env operations for security.",
		expression:     `strenv("MYENV")`,
		expectedError:  "env operations have been disabled",
	},
	{
		description:    "envsubst() operation fails when security is enabled",
		subdescription: "Use `--security-disable-env-ops` to disable env operations for security.",
		expression:     `"value: ${MYENV}" | envsubst`,
		expectedError:  "env operations have been disabled",
	},
}

func TestEnvOperatorSecurityDisabledScenarios(t *testing.T) {
	// Save original security preferences
	originalDisableEnvOps := ConfiguredSecurityPreferences.DisableEnvOps
	defer func() {
		ConfiguredSecurityPreferences.DisableEnvOps = originalDisableEnvOps
	}()

	// Test that env() fails when DisableEnvOps is true
	ConfiguredSecurityPreferences.DisableEnvOps = true

	for _, tt := range envOperatorSecurityDisabledScenarios {
		testScenario(t, &tt)
	}
	appendOperatorDocumentScenario(t, "env-variable-operators", envOperatorSecurityDisabledScenarios)
}
