package yqlib

import (
	"container/list"
	"fmt"
	"os"
	"regexp"
	"strings"

	parse "github.com/a8m/envsubst/parse"
)

type envOpPreferences struct {
	StringValue bool
	NoUnset     bool
	NoEmpty     bool
	FailFast    bool
}

var envVarNameRegexp = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func envOperator(d *dataTreeNavigator, context Context, expressionNode *ExpressionNode) (Context, error) {
	if ConfiguredSecurityPreferences.DisableEnvOps {
		return Context{}, fmt.Errorf("env operations have been disabled")
	}
	envNameExp := expressionNode.Operation.CandidateNode.Value
	log.Debugf("EnvOperator, env expression: %v", envNameExp)

	preferences := expressionNode.Operation.Preferences.(envOpPreferences)

	if envVarNameRegexp.MatchString(envNameExp) {
		if rawValue, ok := os.LookupEnv(envNameExp); ok {
			node, err := envValueToNode(envNameExp, rawValue, preferences)
			if err != nil {
				return Context{}, err
			}
			return context.SingleChildContext(node), nil
		}
	}

	return dynamicEnvOperator(d, context, envNameExp, preferences)
}

// dynamicEnvOperator handles the case where the text between the brackets of env()/strenv()
// is not a literal, already-set environment variable name - instead it's treated as a yq
// expression that is evaluated against each candidate to work out which variable to read.
func dynamicEnvOperator(d *dataTreeNavigator, context Context, envNameExp string, preferences envOpPreferences) (Context, error) {
	nameExpNode, err := ExpressionParser.ParseExpression(envNameExp)
	if err != nil {
		return Context{}, fmt.Errorf("could not process substitution '%v' in env(): %w", envNameExp, err)
	}

	results := list.New()

	for el := context.MatchingNodes.Front(); el != nil; el = el.Next() {
		candidate := el.Value.(*CandidateNode)

		nameResults, err := d.GetMatchingNodes(context.SingleReadonlyChildContext(candidate), nameExpNode)
		if err != nil {
			return Context{}, err
		}

		if nameResults.MatchingNodes.Len() != 1 {
			return Context{}, fmt.Errorf("substitution '%v' in env() must resolve to exactly one value, found %v", envNameExp, nameResults.MatchingNodes.Len())
		}

		nameNode := nameResults.MatchingNodes.Front().Value.(*CandidateNode)
		if nameNode.Kind != ScalarNode {
			return Context{}, fmt.Errorf("substitution '%v' in env() must resolve to a scalar value, found %v", envNameExp, nameNode.Tag)
		}

		envName := nameNode.Value
		node, err := envValueToNode(envName, os.Getenv(envName), preferences)
		if err != nil {
			return Context{}, err
		}
		results.PushBack(node)
	}

	return context.ChildContext(results), nil
}

func envValueToNode(envName string, rawValue string, preferences envOpPreferences) (*CandidateNode, error) {
	if preferences.StringValue {
		return &CandidateNode{
			Kind:  ScalarNode,
			Tag:   "!!str",
			Value: rawValue,
		}, nil
	}

	if rawValue == "" {
		return nil, fmt.Errorf("value for env variable '%v' not provided in env()", envName)
	}

	decoder := NewYamlDecoder(ConfiguredYamlPreferences)
	if err := decoder.Init(strings.NewReader(rawValue)); err != nil {
		return nil, err
	}
	node, err := decoder.Decode()
	if err != nil {
		return nil, err
	}

	log.Debugf("ENV tag: %v", node.Tag)
	log.Debugf("ENV value: %v", node.Value)
	log.Debugf("ENV Kind: %v", node.Kind)

	return node, nil
}

func envsubstOperator(_ *dataTreeNavigator, context Context, expressionNode *ExpressionNode) (Context, error) {
	if ConfiguredSecurityPreferences.DisableEnvOps {
		return Context{}, fmt.Errorf("env operations have been disabled")
	}
	var results = list.New()
	preferences := envOpPreferences{}
	if expressionNode.Operation.Preferences != nil {
		preferences = expressionNode.Operation.Preferences.(envOpPreferences)
	}

	parser := parse.New("string", os.Environ(),
		&parse.Restrictions{NoUnset: preferences.NoUnset, NoEmpty: preferences.NoEmpty})

	if preferences.FailFast {
		parser.Mode = parse.Quick
	} else {
		parser.Mode = parse.AllErrors
	}

	for el := context.MatchingNodes.Front(); el != nil; el = el.Next() {
		node := el.Value.(*CandidateNode)
		if node.Tag != "!!str" {
			log.Warningf("EnvSubstOperator, env name: %v %v", node.Tag, node.Value)
			return Context{}, fmt.Errorf("cannot substitute with %v, can only substitute strings. Hint: Most often you'll want to use '|=' over '=' for this operation", node.Tag)
		}

		value, err := parser.Parse(node.Value)
		if err != nil {
			return Context{}, err
		}
		result := node.CreateReplacement(ScalarNode, "!!str", value)
		results.PushBack(result)
	}

	return context.ChildContext(results), nil
}
