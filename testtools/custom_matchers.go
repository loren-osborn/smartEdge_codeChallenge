package testtools

import (
	"errors"
	"fmt"
	"github.com/onsi/gomega/types"
	"github.com/xeipuuv/gojsonschema"
	"strings"
)

type conformToJSONSchemaFileMatcher struct {
	schemaFileURL string
}

type resultOfJSONSchemaMatch struct {
	matcher          *conformToJSONSchemaFileMatcher
	actualJSONString string
	docLoader        gojsonschema.JSONLoader
	schemaLoader     gojsonschema.JSONLoader
	result           *gojsonschema.Result
	err              error
}

// internalMatch contains all the shared logic from all of
// conformToJSONSchemaFileMatcher's methods
func (sfm *conformToJSONSchemaFileMatcher) internalMatch(actual interface{}) *resultOfJSONSchemaMatch {
	result := resultOfJSONSchemaMatch{
		matcher: sfm,
	}
	result.schemaLoader = gojsonschema.NewReferenceLoader(sfm.schemaFileURL)
	var ok bool
	if result.actualJSONString, ok = actual.(string); !ok {
		result.err = errors.New("actual must be a string")
		return &result
	}
	result.docLoader = gojsonschema.NewStringLoader(result.actualJSONString)

	result.result, result.err = gojsonschema.Validate(result.schemaLoader, result.docLoader)
	return &result
}

// ConformToJSONSchemaFile returns a Gomega "custom matcher" to validate a JSON
// blob against a JSON schema
func ConformToJSONSchemaFile(schemaURL string) types.GomegaMatcher {
	return &conformToJSONSchemaFileMatcher{
		schemaFileURL: schemaURL,
	}
}

// Match returns whether actual conforms to the provided schema, and returns a
// non-nil error if there is an issue reading or parsing the provided schema
func (sfm *conformToJSONSchemaFileMatcher) Match(actual interface{}) (success bool, err error) {
	result := sfm.internalMatch(actual)
	if result.err != nil {
		return false, result.err
	}
	return result.result.Valid(), nil
}

// FailureMessage describes why actual failed to conform to the
// provided schema
func (sfm *conformToJSONSchemaFileMatcher) FailureMessage(actual interface{}) (message string) {
	actualDisplay := " a non-string "
	result := sfm.internalMatch(actual)
	if (result.err == nil) || (result.actualJSONString != "") {
		actualDisplay = fmt.Sprintf("\n\t%#v\n", result.actualJSONString)
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Expected%sto conform to the JSON schema in %s but failed because", actualDisplay, sfm.schemaFileURL))
	if result.err != nil {
		sb.WriteString(" ")
		sb.WriteString(result.err.Error())
	} else if result.result != nil {
		sb.WriteString(":")
		for _, desc := range result.result.Errors() {
			sb.WriteString("\n\t- ")
			sb.WriteString(desc.String())
		}
	}
	return sb.String()
}

// NegatedFailureMessage describes why actual unexpectedly conformed to the
// provided schema
func (sfm *conformToJSONSchemaFileMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	actualDisplay := "a non-string"
	result := sfm.internalMatch(actual)
	if (result.err == nil) || (result.actualJSONString != "") {
		actualDisplay = fmt.Sprintf("%#v", result.actualJSONString)
	}
	return fmt.Sprintf("Expected\n\t%s\nnot to conform to the JSON schema in %s", actualDisplay, sfm.schemaFileURL)
}
