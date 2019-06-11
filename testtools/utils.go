package testtools

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
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
	MatchString(string) bool
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
func (ssm *StringStringMatcher) MatchString(s string) bool {
	return ssm.str == s
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

// MatchString returns whether the strings are equal
func (rsm *RegexpStringMatcher) MatchString(s string) bool {
	return rsm.re.MatchString(s)
}

// MatchString returns a printable representation of the regular expression
func (rsm *RegexpStringMatcher) String() string {
	return fmt.Sprintf("regexp.Regexp(%#v)", rsm.re.String())
}
