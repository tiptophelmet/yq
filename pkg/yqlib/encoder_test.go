//go:build !yq_nojson

package yqlib

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/mikefarah/yq/v4/test"
)

func yamlToJSON(t *testing.T, sampleYaml string, indent int) string {
	t.Helper()
	var output bytes.Buffer
	writer := bufio.NewWriter(&output)

	prefs := ConfiguredJSONPreferences.Copy()
	prefs.Indent = indent
	prefs.UnwrapScalar = false
	var jsonEncoder = NewJSONEncoder(prefs)
	inputs, err := readDocuments(strings.NewReader(sampleYaml), "sample.yml", 0, NewYamlDecoder(ConfiguredYamlPreferences))
	if err != nil {
		panic(err)
	}
	node := inputs.Front().Value.(*CandidateNode)
	log.Debugf("%v", NodeToString(node))
	// log.Debugf("Content[0] %v", NodeToString(node.Content[0]))

	err = jsonEncoder.Encode(writer, node)
	if err != nil {
		panic(err)
	}
	writer.Flush()

	return strings.TrimSuffix(output.String(), "\n")
}

func TestJSONEncoderPreservesObjectOrder(t *testing.T) {
	var sampleYaml = `zabbix: winner
apple: great
banana:
- {cobra: kai, angus: bob}
`
	var expectedJSON = `{
  "zabbix": "winner",
  "apple": "great",
  "banana": [
    {
      "cobra": "kai",
      "angus": "bob"
    }
  ]
}`
	var actualJSON = yamlToJSON(t, sampleYaml, 2)
	test.AssertResult(t, expectedJSON, actualJSON)
}

func TestJsonNullInArray(t *testing.T) {
	var sampleYaml = `[null]`
	var actualJSON = yamlToJSON(t, sampleYaml, 0)
	test.AssertResult(t, sampleYaml, actualJSON)
}

func TestJsonNull(t *testing.T) {
	var sampleYaml = `null`
	var actualJSON = yamlToJSON(t, sampleYaml, 0)
	test.AssertResult(t, sampleYaml, actualJSON)
}

func TestJsonNullInObject(t *testing.T) {
	var sampleYaml = `{x: null}`
	var actualJSON = yamlToJSON(t, sampleYaml, 0)
	test.AssertResult(t, `{"x":null}`, actualJSON)
}

func TestJsonEncoderDoesNotEscapeHTMLChars(t *testing.T) {
	var sampleYaml = `build: "( ./lint && ./format && ./compile ) < src.code"`
	var expectedJSON = `{"build":"( ./lint && ./format && ./compile ) < src.code"}`
	var actualJSON = yamlToJSON(t, sampleYaml, 0)
	test.AssertResult(t, expectedJSON, actualJSON)
}

func yamlToYaml(t *testing.T, sampleYaml string) string {
	t.Helper()
	var output bytes.Buffer
	writer := bufio.NewWriter(&output)

	prefs := ConfiguredYamlPreferences.Copy()
	prefs.UnwrapScalar = false
	var yamlEncoder = NewYamlEncoder(prefs)
	inputs, err := readDocuments(strings.NewReader(sampleYaml), "sample.yml", 0, NewYamlDecoder(ConfiguredYamlPreferences))
	if err != nil {
		panic(err)
	}
	node := inputs.Front().Value.(*CandidateNode)

	err = yamlEncoder.Encode(writer, node)
	if err != nil {
		panic(err)
	}
	writer.Flush()

	return output.String()
}

func TestYamlEncoderTopLevelSingleQuotedClosingIndent(t *testing.T) {
	var sampleYaml = "'Get quacked.\n\n- Duck\n\n'\n"
	var expected = "'Get quacked.\n\n  - Duck\n\n  '\n"
	var actual = yamlToYaml(t, sampleYaml)
	test.AssertResult(t, expected, actual)
}

func TestYamlEncoderFoldedBlockNoPhantomBlankLines(t *testing.T) {
	var sampleYaml = `env:
  YQ_QUERY: >-
    (.runs.steps // .jobs.*.steps) | map(
      select(.uses | test("@(v\d+(\.\d+)*|\d+(\.\d+)+|[a-f0-9]{40})$") | not)
      | .uses
      | "\(key | line) \(split(\"@\")[1])"
    )[]
`
	var actual = yamlToYaml(t, sampleYaml)
	test.AssertResult(t, sampleYaml, actual)
}

func TestYamlEncoderFoldedBlockPreservesGenuineBlankLine(t *testing.T) {
	var sampleYaml = "key: >-\n  line1\n\n  line2\n"
	var actual = yamlToYaml(t, sampleYaml)
	test.AssertResult(t, sampleYaml, actual)
}

func TestYamlEncoderFoldedBlockSingleLineUnaffected(t *testing.T) {
	var sampleYaml = "key: >-\n  single line, no newlines here\n"
	var actual = yamlToYaml(t, sampleYaml)
	test.AssertResult(t, sampleYaml, actual)
}

func TestYamlEncoderLiteralBlockUnaffected(t *testing.T) {
	var sampleYaml = `env:
  YQ_QUERY: |-
    (.runs.steps // .jobs.*.steps) | map(
      select(.uses | test("@(v\d+(\.\d+)*|\d+(\.\d+)+|[a-f0-9]{40})$") | not)
      | .uses
      | "\(key | line) \(split(\"@\")[1])"
    )[]
`
	var actual = yamlToYaml(t, sampleYaml)
	test.AssertResult(t, sampleYaml, actual)
}
