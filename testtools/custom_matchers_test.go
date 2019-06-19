package testtools_test

import (
	"fmt"
	"github.com/onsi/gomega/types"
	"github.com/smartedge/codechallenge/deps"
	"github.com/smartedge/codechallenge/testtools"
	"testing"
)

// TestSchemaConformance
func TestSchemaConformance(t *testing.T) {
	badSchemaUrl := "fred:///a/b/c/not%20a%20file%20name/12345"
	var matcher types.GomegaMatcher = testtools.ConformToJSONSchemaFile(
		badSchemaUrl)
	expectedNotStringErr := &testtools.ErrorSpec{
		Type:    "*errors.errorString",
		Message: "actual must be a string",
	}
	if matches, err := matcher.Match(nil); matches {
		t.Error(
			"ConformToJSONSchemaFile with actual a non-string should return false")
	} else if err2 := expectedNotStringErr.EnsureMatches(err); err2 != nil {
		t.Error(err2.Error())
	}
	expectedErrorFailureMsg := fmt.Sprintf(
		"Expected a non-string to conform to the JSON schema in %s but failed because actual must be a string",
		badSchemaUrl)
	actualErrorFailureMsg := matcher.FailureMessage(nil)
	if actualErrorFailureMsg != expectedErrorFailureMsg {
		t.Errorf(
			"Expected ConformToJSONSchemaFile(%#v).FailureMessage(nil) to return\n\t%#v but got \n\t%#v instead",
			badSchemaUrl, expectedErrorFailureMsg, actualErrorFailureMsg)
	}
	expectedBadPathErr := &testtools.ErrorSpec{
		Type:    "*errors.errorString",
		Message: "Reference fred:///a/b/c/not%20a%20file%20name/12345 must be canonical",
	}
	if matches, err := matcher.Match(""); matches {
		t.Error(
			"ConformToJSONSchemaFile with bad filename should return false")
	} else if err2 := expectedBadPathErr.EnsureMatches(err); err2 != nil {
		t.Error(err2.Error())
	}
	JSONValidationSchemaURL, err := testtools.GetURLFromProjectPath(deps.Defaults, testtools.JSONValidationSchemaPath)
	if err != nil {
		t.Error(err.Error())
	}
	matcher = testtools.ConformToJSONSchemaFile(JSONValidationSchemaURL)
	if matches, err := matcher.Match("{}"); err != nil {
		t.Error(err.Error())
	} else if matches {
		t.Error(
			"ConformToJSONSchemaFile that doesn't conform to schema shouldn't match")
	}
	expectedNegatedFailureMsg := fmt.Sprintf(
		"Expected\n\t\"{}\"\nnot to conform to the JSON schema in %s",
		JSONValidationSchemaURL)
	actualNegatedFailureMsg := matcher.NegatedFailureMessage("{}")
	if actualNegatedFailureMsg != expectedNegatedFailureMsg {
		t.Errorf(
			"Expected ConformToJSONSchemaFile(%#v).NegatedFailureMessage(\"{}\") to return\n\t%#v but got \n\t%#v instead",
			JSONValidationSchemaURL, expectedNegatedFailureMsg, actualNegatedFailureMsg)
	}
	expectedFailureMsg := fmt.Sprintf(
		"Expected\n\t\"{}\"\nto conform to the JSON schema in %s but failed because:\n\t- (root): message is required\n\t- (root): signature is required\n\t- (root): pubkey is required",
		JSONValidationSchemaURL)
	actualFailureMsg := matcher.FailureMessage("{}")
	if actualFailureMsg != expectedFailureMsg {
		t.Errorf(
			"Expected ConformToJSONSchemaFile(%#v).FailureMessage(\"{}\") to return\n\t%#v but got \n\t%#v instead",
			JSONValidationSchemaURL, expectedFailureMsg, actualFailureMsg)
	}
}
