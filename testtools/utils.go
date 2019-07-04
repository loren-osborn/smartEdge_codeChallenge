package testtools

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/smartedge/codechallenge"
	"github.com/smartedge/codechallenge/crypt"
	"github.com/smartedge/codechallenge/deps"
	"github.com/xeipuuv/gojsonschema"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// JSONValidationSchemaPath the path in the project to the valid response schema.
const (
	JSONValidationSchemaPath = "testdata/valid_output_schema.json"
)

// AreFuncsEqual returns true only if a and b are both functions, and
// both point to the same function. Returns false and a non-nil error if either
// argument is not a function. Returns true and a non-nil error if both
// arguments are nil,
func AreFuncsEqual(a interface{}, b interface{}) (bool, error) {
	checkTwoVals := func(matcher func(int) bool, matchDesc string, expectedDesc string) error {
		matches := 0
		for i := 0; i < 2; i++ {
			if matcher(i) {
				matches++
			}
		}
		if matches > 0 {
			errMsg := fmt.Sprintf("Both values %s when %s expected", matchDesc, expectedDesc)
			if matches == 1 {
				which := "Second"
				if matcher(0) {
					which = "First"
				}
				errMsg = fmt.Sprintf("%s value %s when two %s expected", which, matchDesc, expectedDesc)
			}
			return errors.New(errMsg)
		}
		return nil
	}

	valueInfos := [2]struct {
		val             interface{}
		isNil           bool
		valueReflection reflect.Value
		typeReflection  reflect.Type
	}{}
	valueInfos[0].val = a
	valueInfos[1].val = b
	for i := 0; i < 2; i++ {
		valueInfos[i].isNil = valueInfos[i].val == nil
		if !valueInfos[i].isNil {
			valueInfos[i].valueReflection = reflect.ValueOf(valueInfos[i].val)
			valueInfos[i].isNil = valueInfos[i].valueReflection.IsNil()
			valueInfos[i].typeReflection = valueInfos[i].valueReflection.Type()
		}
	}
	if err := checkTwoVals(func(i int) bool { return valueInfos[i].isNil }, "nil", "funcs"); err != nil {
		result := valueInfos[0].isNil && valueInfos[1].isNil && (valueInfos[0].typeReflection == valueInfos[1].typeReflection)
		return result, err
	}
	if err := checkTwoVals(func(i int) bool { return valueInfos[i].typeReflection.Kind() != reflect.Func }, "not a func", "funcs"); err != nil {
		return false, err
	}
	return (valueInfos[0].valueReflection.Pointer() == valueInfos[1].valueReflection.Pointer()), nil
}

// AreStringSlicesEqual determines if two string slices are equal. Equality
// distinguishes nil-ness, but not capacity
func AreStringSlicesEqual(a []string, b []string) bool {
	if (a == nil) || (b == nil) {
		return (a == nil) && (b == nil)
	}
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if b[i] != v {
			return false
		}
	}
	return true
}

// CloneStringSlice create a non-shared copy of inSlice
func CloneStringSlice(inSlice []string) []string {
	if inSlice == nil {
		return nil
	}
	result := make([]string, len(inSlice), cap(inSlice))
	for i, v := range inSlice {
		result[i] = v
	}
	return result
}

// WrapFuncCallWithCounter wraps the provided function, adding a returned
// pointer to a call counter.
func WrapFuncCallWithCounter(f func()) (func(), *int) {
	counter := 0
	wrapped := func() {
		f()
		counter++
	}
	return wrapped, &counter
}

// StringMatcher is an interface for comparing strings.
type StringMatcher interface {
	MatchString(string) error
	String() string
}

// StringStringMatcher returns StringMatcher only matching itself.
type StringStringMatcher struct {
	str string
}

// NewStringStringMatcher returns a new StringStringMatcher from the string s.
func NewStringStringMatcher(s string) *StringStringMatcher {
	return &StringStringMatcher{str: s}
}

// MatchString returns whether the strings are equal
func (ssm *StringStringMatcher) MatchString(s string) error {
	if ssm.str == s {
		return nil
	}
	return fmt.Errorf("Didn't equal string %s", ssm.String())
}

// String returns a printable representation of the string
func (ssm *StringStringMatcher) String() string {
	return fmt.Sprintf("%#v", ssm.str)
}

// RegexpStringMatcher is a thin wrapper over regexp.Regexp to alter its
// String() result.
type RegexpStringMatcher struct {
	re *regexp.Regexp
}

// NewRegexpStringMatcher returns a RegexpStringMatcher from the pattern.
func NewRegexpStringMatcher(pattern string) *RegexpStringMatcher {
	return &RegexpStringMatcher{re: regexp.MustCompile(pattern)}
}

// MatchString returns whether the strings match the pattern.
func (rsm *RegexpStringMatcher) MatchString(s string) error {
	if rsm.re.MatchString(s) {
		return nil
	}
	return fmt.Errorf("Didn't match pattern %s", rsm.String())
}

// String returns a printable representation of the regular expression
func (rsm *RegexpStringMatcher) String() string {
	return fmt.Sprintf("regexp.Regexp(%#v)", rsm.re.String())
}

// GenericStringMatcher is generic matcher for a string validation function.
type GenericStringMatcher func(string) error

// MatchString returns whether the strings match.
func (gsm GenericStringMatcher) MatchString(s string) error {
	return gsm(s)
}

// String returns a generic printable representation
func (gsm GenericStringMatcher) String() string {
	return "validation function"
}

// GetResponseMatcherForMessageAndPubKey returns a string matcher for a valid
// response with a specific message and public key
func GetResponseMatcherForMessageAndPubKey(d *deps.Dependencies, msg string, pubKey string) GenericStringMatcher {
	return GenericStringMatcher(func(r string) error {
		expectedAlg, err := GetPEMPublicKeyAlgorithm(pubKey)
		if err != nil {
			return fmt.Errorf("Invalid key to match: %s", err.Error())
		}
		resp, alg, err := ValidateResponse(d, r)
		if err != nil {
			return err
		}
		if resp.Message != msg {
			return fmt.Errorf("signed message was %#v when %#v was expected", resp.Message, msg)
		}
		if alg != expectedAlg {
			return fmt.Errorf("signed with %s when %s expected", alg.String(), expectedAlg.String())
		}
		if resp.Pubkey != pubKey {
			return fmt.Errorf("unexpected key generated by fixed pseudo-random entropy:\n%s", resp.Pubkey)
		}
		return nil
	})
}

// GetResponseMatcherForMessageAndAlgorithm returns a string matcher for a valid
// response with a specific message and algorithm
func GetResponseMatcherForMessageAndAlgorithm(d *deps.Dependencies, msg string, algorithm x509.PublicKeyAlgorithm) GenericStringMatcher {
	return GenericStringMatcher(func(r string) error {
		resp, alg, err := ValidateResponse(d, r)
		if err != nil {
			return err
		}
		if resp.Message != msg {
			return fmt.Errorf("signed message was %#v when %#v was expected", resp.Message, msg)
		}
		if alg != algorithm {
			return fmt.Errorf("signed with %s when %s expected", alg.String(), algorithm.String())
		}
		return nil
	})
}

// GetPEMPublicKeyAlgorithm determines the PKI algorithm from a given public
// key string. (When the production code supports verification, this will be
// moved to a methods of the X509Encoded and PEMEncoded types)
func GetPEMPublicKeyAlgorithm(PEMPublicKey string) (x509.PublicKeyAlgorithm, error) {
	keyType := x509.UnknownPublicKeyAlgorithm
	pubKeyPEM := crypt.NewPEMBufferFromString(PEMPublicKey)
	pubKeyX509, err := pubKeyPEM.DecodeToX509()
	if err != nil {
		return keyType, err
	}
	untypedPubKey, err := pubKeyX509.AsGenericPublicKey()
	if err != nil {
		return keyType, err
	}
	switch untypedPubKey.(type) {
	case *ecdsa.PublicKey:
		keyType = x509.ECDSA
	case *rsa.PublicKey:
		keyType = x509.RSA
	default:
		return keyType, fmt.Errorf("Public key did not conform to recognized algorithm:\n%s", PEMPublicKey)
	}
	return keyType, nil
}

// GetURLFromProjectPath converts a project path to a file:/// URL.
func GetURLFromProjectPath(d *deps.Dependencies, projPath string) (string, error) {
	projectRoot, err := GetProjectPath(d)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("file://%s/%s", projectRoot, projPath), nil
}

// ValidateJSONSchema validates that a string containing JSON conforms to
// the specified schema.
func ValidateJSONSchema(d *deps.Dependencies, responseStr string, schemaPath string) error {
	JSONValidationSchemaURL, err := GetURLFromProjectPath(d, schemaPath)
	if err != nil {
		return err
	}
	schemaLoader := gojsonschema.NewReferenceLoader(JSONValidationSchemaURL)
	docLoader := gojsonschema.NewStringLoader(responseStr)
	matchedSchema, err := gojsonschema.Validate(schemaLoader, docLoader)
	if (err == nil) && !(matchedSchema.Valid()) {
		errMsgBuilder := &strings.Builder{}
		for _, desc := range matchedSchema.Errors() {
			fmt.Fprintf(errMsgBuilder, "\n- %s", desc)
		}
		return fmt.Errorf("Response didn't conform to JSON schema in %s%s", JSONValidationSchemaURL, errMsgBuilder.String())
	}
	return err
}

// ValidateResponse validates that the string provided conforms to the
// schema in the spec, has a valid PEM format public key, a valid base64
// signature, and that the signature was produced by signing the message with
// the private key (corresponding to the included public key). It returns the
// parsed JSON response, the signing algorithm and any errors encountered.
func ValidateResponse(d *deps.Dependencies, responseStr string) (*codechallenge.SignedMessage, x509.PublicKeyAlgorithm, error) {
	keyType := x509.UnknownPublicKeyAlgorithm
	if err := ValidateJSONSchema(d, responseStr, JSONValidationSchemaPath); err != nil {
		return nil, keyType, err
	}
	response := codechallenge.SignedMessage{}
	if err := json.Unmarshal([]byte(responseStr), &response); err != nil {
		return nil, keyType, err
	}
	keyType, err := GetPEMPublicKeyAlgorithm(response.Pubkey)
	if err != nil {
		return &response, keyType, err
	}
	settings := &crypt.PkiSettings{
		Algorithm:      keyType,
		RSAKeyBits:     2048,
		PrivateKeyPath: "",
		PublicKeyPath:  "",
	}
	tooling, err := crypt.GetCryptoTooling(nil, settings)
	if err != nil {
		return &response, keyType, err
	}
	matched, err := tooling.VerifySignedMessage(response.Message, response.Signature, response.Pubkey)
	if (err == nil) && !matched {
		return &response, keyType, fmt.Errorf("Code signature invalid for message: %#v", response.Message)
	}
	return &response, keyType, err
}

// GetSourceFilename returns the filename of the caller's source code.
func GetSourceFilename(d *deps.Dependencies) (string, error) {
	// From https://stackoverflow.com/questions/47218715/is-it-possible-to-get-filename-where-code-is-called-in-golang
	_, file, _, ok := d.Runtime.Caller(1)
	if !ok {
		return "", errors.New("unable to get source code filename")
	}
	return filepath.Clean(file), nil
}

// GetProjectPath returns the root path of the project (at compile time.)
func GetProjectPath(d *deps.Dependencies) (string, error) {
	file, err := GetSourceFilename(d)
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Dir(filepath.Dir(file))), nil
}

// StringPtr returns a pointer to a newly allocated copy of the string.
func StringPtr(s string) *string {
	return &s
}

// AreFakeFileSystemsEqual does a deep comparison of two maps of strings to
// string pointers (being used as fake filesystems.)
func AreFakeFileSystemsEqual(a map[string]*string, b map[string]*string) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	for ak, avp := range a {
		bvp, ok := b[ak]
		if !ok || ((avp == nil) != (bvp == nil)) {
			return false
		}
		if avp != nil {
			if *avp != *bvp {
				return false
			}
		}
	}
	for bk := range b {
		_, ok := a[bk]
		if !ok {
			return false
		}
	}
	return true
}

// CloneFakeFileSystemsEqual makes a deep clone of the map.
func CloneFakeFileSystemsEqual(orig map[string]*string) map[string]*string {
	if orig == nil {
		return nil
	}
	result := make(map[string]*string, len(orig))
	for key, val := range orig {
		if val == nil {
			result[key] = nil
		} else {
			result[key] = StringPtr(*val)
		}
	}
	return result
}

// DummyFileInfo a mock for a os.FileInfo interface.
type DummyFileInfo struct {
	NameVal    string
	SizeVal    int64
	ModeVal    os.FileMode
	ModTimeVal time.Time
}

// Name base name of the file
func (dfi *DummyFileInfo) Name() string {
	return dfi.NameVal
}

// Size length in bytes for regular files; system-dependent for others
func (dfi *DummyFileInfo) Size() int64 {
	return dfi.SizeVal
}

// Mode file mode bits
func (dfi *DummyFileInfo) Mode() os.FileMode {
	return dfi.ModeVal
}

// ModTime modification time
func (dfi *DummyFileInfo) ModTime() time.Time {
	return dfi.ModTimeVal
}

// IsDir abbreviation for Mode().IsDir()
func (dfi *DummyFileInfo) IsDir() bool {
	return dfi.Mode().IsDir()
}

// Sys underlying data source (can return nil)
func (dfi *DummyFileInfo) Sys() interface{} {
	return nil
}
