package httpexpect

import (
	"errors"
	"fmt"
	"reflect"
)

// Array provides methods to inspect attached []interface{} object
// (Go representation of JSON array).
type Array struct {
	chain *chain
	value []interface{}
}

// NewArray returns a new Array instance.
//
// Both reporter and value should not be nil. If value is nil, failure is
// reported.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
func NewArray(reporter Reporter, value []interface{}) *Array {
	return newArray(newChainWithDefaults("Array()", reporter), value)
}

func newArray(parent *chain, val []interface{}) *Array {
	a := &Array{parent.clone(), nil}

	if val == nil {
		a.chain.fail(AssertionFailure{
			Type:   AssertNotNil,
			Actual: &AssertionValue{val},
			Errors: []error{
				errors.New("expected: non-nil array"),
			},
		})
	} else {
		a.value, _ = canonArray(a.chain, val)
	}

	return a
}

// Raw returns underlying value attached to Array.
// This is the value originally passed to NewArray, converted to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	assert.Equal(t, []interface{}{"foo", 123.0}, array.Raw())
func (a *Array) Raw() []interface{} {
	return a.value
}

// Path is similar to Value.Path.
func (a *Array) Path(path string) *Value {
	a.chain.enter("Path(%q)", path)
	defer a.chain.leave()

	return jsonPath(a.chain, a.value, path)
}

// Schema is similar to Value.Schema.
func (a *Array) Schema(schema interface{}) *Array {
	a.chain.enter("Schema()")
	defer a.chain.leave()

	jsonSchema(a.chain, a.value, schema)
	return a
}

// Length returns a new Number instance with array length.
//
// Example:
//
//	array := NewArray(t, []interface{}{1, 2, 3})
//	array.Length().Equal(3)
func (a *Array) Length() *Number {
	a.chain.enter("Length()")
	defer a.chain.leave()

	if a.chain.failed() {
		return newNumber(a.chain, 0)
	}

	return newNumber(a.chain, float64(len(a.value)))
}

// Element returns a new Value instance with array element for given index.
//
// If index is out of array bounds, Element reports failure and returns empty
// (but non-nil) instance.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.Element(0).String().Equal("foo")
//	array.Element(1).Number().Equal(123)
func (a *Array) Element(index int) *Value {
	a.chain.enter("Element(%d)", index)
	defer a.chain.leave()

	if a.chain.failed() {
		return newValue(a.chain, nil)
	}

	if index < 0 || index >= len(a.value) {
		a.chain.fail(AssertionFailure{
			Type:   AssertInRange,
			Actual: &AssertionValue{index},
			Expected: &AssertionValue{AssertionRange{
				Min: 0,
				Max: len(a.value) - 1,
			}},
			Errors: []error{
				errors.New("expected: valid element index"),
			},
		})
		return newValue(a.chain, nil)
	}

	return newValue(a.chain, a.value[index])
}

// First returns a new Value instance for the first element of array.
//
// If given array is empty, First reports failure and returns empty
// (but non-nil) instance.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.First().String().Equal("foo")
func (a *Array) First() *Value {
	a.chain.enter("First()")
	defer a.chain.leave()

	if a.chain.failed() {
		return newValue(a.chain, nil)
	}

	if len(a.value) == 0 {
		a.chain.fail(AssertionFailure{
			Type:   AssertNotEmpty,
			Actual: &AssertionValue{a.value},
			Errors: []error{
				errors.New("expected: non-empty array"),
			},
		})
		return newValue(a.chain, nil)
	}

	return newValue(a.chain, a.value[0])
}

// Last returns a new Value instance for the last element of array.
//
// If given array is empty, Last reports failure and returns empty
// (but non-nil) instance.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.Last().Number().Equal(123)
func (a *Array) Last() *Value {
	a.chain.enter("Last()")
	defer a.chain.leave()

	if a.chain.failed() {
		return newValue(a.chain, nil)
	}

	if len(a.value) == 0 {
		a.chain.fail(AssertionFailure{
			Type:   AssertNotEmpty,
			Actual: &AssertionValue{a.value},
			Errors: []error{
				errors.New("expected: non-empty array"),
			},
		})
		return newValue(a.chain, nil)
	}

	return newValue(a.chain, a.value[len(a.value)-1])
}

// Iter returns a new slice of Values attached to array elements.
//
// Example:
//
//	strings := []interface{}{"foo", "bar"}
//	array := NewArray(t, strings)
//
//	for n, val := range array.Iter() {
//	    val.String().Equal(strings[n])
//	}
func (a *Array) Iter() []Value {
	if a.chain.failed() {
		return []Value{}
	}

	ret := []Value{}
	for n := range a.value {
		valueChain := a.chain.clone()
		valueChain.enter("Iter[%d]", n)

		ret = append(ret, *newValue(valueChain, a.value[n]))
	}

	return ret
}

// Every runs the passed function on all the Elements in the array.
//
// If assertion inside function fails, the original Array is marked failed.
//
// Every will execute the function for all values in the array irrespective
// of assertion failures for some values in the array.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", "bar"})
//
//	array.Every(func(index int, value *httpexpect.Value) {
//		value.String().NotEmpty()
//	})
func (a *Array) Every(fn func(index int, value *Value)) *Array {
	a.chain.enter("Every()")
	defer a.chain.leave()

	if a.chain.failed() {
		return a
	}

	if fn == nil {
		a.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil function argument"),
			},
		})
		return a
	}

	chainFailure := false

	for index, val := range a.value {
		valueChain := a.chain.clone()
		valueChain.replace("Every[%d]", index)

		valueChain.setFailCallback(func() {
			chainFailure = true
		})

		fn(index, newValue(valueChain, val))
	}

	if chainFailure {
		a.chain.setFailed()
	}

	return a
}

// Empty succeeds if array is empty.
//
// Example:
//
//	array := NewArray(t, []interface{}{})
//	array.Empty()
func (a *Array) Empty() *Array {
	a.chain.enter("Empty()")
	defer a.chain.leave()

	if a.chain.failed() {
		return a
	}

	if !(len(a.value) == 0) {
		a.chain.fail(AssertionFailure{
			Type:   AssertEmpty,
			Actual: &AssertionValue{a.value},
			Errors: []error{
				errors.New("expected: empty array"),
			},
		})
	}

	return a
}

// NotEmpty succeeds if array is non-empty.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.NotEmpty()
func (a *Array) NotEmpty() *Array {
	a.chain.enter("NotEmpty()")
	defer a.chain.leave()

	if a.chain.failed() {
		return a
	}

	if !(len(a.value) != 0) {
		a.chain.fail(AssertionFailure{
			Type:   AssertNotEmpty,
			Actual: &AssertionValue{a.value},
			Errors: []error{
				errors.New("expected: non-empty array"),
			},
		})
	}

	return a
}

// Equal succeeds if array is equal to given value.
// Before comparison, both array and value are converted to canonical form.
//
// value should be a slice of any type.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.Equal([]interface{}{"foo", 123})
//
//	array := NewArray(t, []interface{}{"foo", "bar"})
//	array.Equal([]string{}{"foo", "bar"})
//
//	array := NewArray(t, []interface{}{123, 456})
//	array.Equal([]int{}{123, 456})
func (a *Array) Equal(value interface{}) *Array {
	a.chain.enter("Equal()")
	defer a.chain.leave()

	if a.chain.failed() {
		return a
	}

	expected, ok := canonArray(a.chain, value)
	if !ok {
		return a
	}

	if !reflect.DeepEqual(expected, a.value) {
		a.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{a.value},
			Expected: &AssertionValue{expected},
			Errors: []error{
				errors.New("expected: arrays are equal"),
			},
		})
	}

	return a
}

// NotEqual succeeds if array is not equal to given value.
// Before comparison, both array and value are converted to canonical form.
//
// value should be a slice of any type.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.NotEqual([]interface{}{123, "foo"})
func (a *Array) NotEqual(value interface{}) *Array {
	a.chain.enter("NotEqual()")
	defer a.chain.leave()

	if a.chain.failed() {
		return a
	}

	expected, ok := canonArray(a.chain, value)
	if !ok {
		return a
	}

	if reflect.DeepEqual(expected, a.value) {
		a.chain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{a.value},
			Expected: &AssertionValue{expected},
			Errors: []error{
				errors.New("expected: arrays are non-equal"),
			},
		})
	}

	return a
}

// EqualUnordered succeeds if array is equal to another array, ignoring element
// order. Before comparison, both arrays are converted to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.EqualUnordered([]interface{}{123, "foo"})
func (a *Array) EqualUnordered(value interface{}) *Array {
	a.chain.enter("EqualUnordered()")
	defer a.chain.leave()

	if a.chain.failed() {
		return a
	}

	expected, ok := canonArray(a.chain, value)
	if !ok {
		return a
	}

	for _, element := range expected {
		expectedCount := countElement(expected, element)
		actualCount := countElement(a.value, element)

		if actualCount != expectedCount {
			if expectedCount == 1 && actualCount == 0 {
				a.chain.fail(AssertionFailure{
					Type:      AssertContainsElement,
					Actual:    &AssertionValue{a.value},
					Expected:  &AssertionValue{element},
					Reference: &AssertionValue{value},
					Errors: []error{
						errors.New("expected: array contains element from reference array"),
					},
				})
			} else {
				a.chain.fail(AssertionFailure{
					Type:      AssertNotContainsElement,
					Actual:    &AssertionValue{a.value},
					Expected:  &AssertionValue{element},
					Reference: &AssertionValue{value},
					Errors: []error{
						fmt.Errorf(
							"expected: element occurs %d time(s), as in reference array,"+
								" but it occurs %d time(s)",
							expectedCount,
							actualCount),
					},
				})
			}
			return a
		}
	}

	for _, element := range a.value {
		expectedCount := countElement(expected, element)
		actualCount := countElement(a.value, element)

		if actualCount != expectedCount {
			if expectedCount == 0 && actualCount == 1 {
				a.chain.fail(AssertionFailure{
					Type:      AssertNotContainsElement,
					Actual:    &AssertionValue{a.value},
					Expected:  &AssertionValue{element},
					Reference: &AssertionValue{value},
					Errors: []error{
						errors.New("expected: array does not contain elements" +
							" that are not present in reference array"),
					},
				})
			} else {
				a.chain.fail(AssertionFailure{
					Type:      AssertNotContainsElement,
					Actual:    &AssertionValue{a.value},
					Expected:  &AssertionValue{element},
					Reference: &AssertionValue{value},
					Errors: []error{
						fmt.Errorf(
							"expected: element occurs %d time(s), as in reference array,"+
								" but it occurs %d time(s)",
							expectedCount,
							actualCount),
					},
				})
			}
			return a
		}
	}

	return a
}

// NotEqualUnordered succeeds if array is not equal to another array, ignoring
// element order. Before comparison, both arrays are converted to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.NotEqualUnordered([]interface{}{123, "foo", "bar"})
func (a *Array) NotEqualUnordered(value interface{}) *Array {
	a.chain.enter("NotEqualUnordered()")
	defer a.chain.leave()

	if a.chain.failed() {
		return a
	}

	expected, ok := canonArray(a.chain, value)
	if !ok {
		return a
	}

	different := false

	for _, element := range expected {
		expectedCount := countElement(expected, element)
		actualCount := countElement(a.value, element)

		if actualCount != expectedCount {
			different = true
			break
		}
	}

	for _, element := range a.value {
		expectedCount := countElement(expected, element)
		actualCount := countElement(a.value, element)

		if actualCount != expectedCount {
			different = true
			break
		}
	}

	if !different {
		a.chain.fail(AssertionFailure{
			Type:      AssertNotEqual,
			Actual:    &AssertionValue{a.value},
			Expected:  &AssertionValue{value},
			Reference: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: arrays are non-equal (ignoring order)"),
			},
		})
	}

	return a
}

// Elements succeeds if array contains all given elements, in given order, and only
// them. Before comparison, array and all elements are converted to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.Elements("foo", 123)
//
// These calls are equivalent:
//
//	array.Elements("a", "b")
//	array.Equal([]interface{}{"a", "b"})
func (a *Array) Elements(values ...interface{}) *Array {
	a.chain.enter("Elements()")
	defer a.chain.leave()

	if a.chain.failed() {
		return a
	}

	expected, ok := canonArray(a.chain, values)
	if !ok {
		return a
	}

	if !reflect.DeepEqual(expected, a.value) {
		a.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{a.value},
			Expected: &AssertionValue{expected},
			Errors: []error{
				errors.New("expected: arrays are equal"),
			},
		})
	}

	return a
}

// NotElements is opposite to Elements.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.NotElements("foo")
//	array.NotElements("foo", 123, 456)
//	array.NotElements(123, "foo")
//
// These calls are equivalent:
//
//	array.NotElements("a", "b")
//	array.NotEqual([]interface{}{"a", "b"})
func (a *Array) NotElements(values ...interface{}) *Array {
	a.chain.enter("Elements()")
	defer a.chain.leave()

	if a.chain.failed() {
		return a
	}

	expected, ok := canonArray(a.chain, values)
	if !ok {
		return a
	}

	if reflect.DeepEqual(expected, a.value) {
		a.chain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{a.value},
			Expected: &AssertionValue{expected},
			Errors: []error{
				errors.New("expected: arrays are non-equal"),
			},
		})
	}

	return a
}

// Contains succeeds if array contains all given elements (in any order).
// Before comparison, array and all elements are converted to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.Contains(123, "foo")
func (a *Array) Contains(values ...interface{}) *Array {
	a.chain.enter("Contains()")
	defer a.chain.leave()

	if a.chain.failed() {
		return a
	}

	elements, ok := canonArray(a.chain, values)
	if !ok {
		return a
	}

	for _, expected := range elements {
		if !(countElement(a.value, expected) != 0) {
			a.chain.fail(AssertionFailure{
				Type:      AssertContainsElement,
				Actual:    &AssertionValue{a.value},
				Expected:  &AssertionValue{expected},
				Reference: &AssertionValue{values},
				Errors: []error{
					errors.New("expected: array contains element from reference array"),
				},
			})
			break
		}
	}

	return a
}

// NotContains succeeds if array contains none of given elements.
// Before comparison, array and all elements are converted to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.NotContains("bar")         // success
//	array.NotContains("bar", "foo")  // failure (array contains "foo")
func (a *Array) NotContains(values ...interface{}) *Array {
	a.chain.enter("NotContains()")
	defer a.chain.leave()

	if a.chain.failed() {
		return a
	}

	elements, ok := canonArray(a.chain, values)
	if !ok {
		return a
	}

	for _, expected := range elements {
		if !(countElement(a.value, expected) == 0) {
			a.chain.fail(AssertionFailure{
				Type:      AssertNotContainsElement,
				Actual:    &AssertionValue{a.value},
				Expected:  &AssertionValue{expected},
				Reference: &AssertionValue{values},
				Errors: []error{
					errors.New("expected:" +
						" array does not contain any elements from reference array"),
				},
			})
			break
		}
	}

	return a
}

// ContainsOnly succeeds if array contains all given elements, in any order, and only
// them, ignoring duplicates. Before comparison, array and all elements are converted
// to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123, 123})
//	array.ContainsOnly(123, "foo")
//
// These calls are equivalent:
//
//	array.ContainsOnly("a", "b")
//	array.ContainsOnly("b", "a")
func (a *Array) ContainsOnly(values ...interface{}) *Array {
	a.chain.enter("ContainsOnly()")
	defer a.chain.leave()

	if a.chain.failed() {
		return a
	}

	elements, ok := canonArray(a.chain, values)
	if !ok {
		return a
	}

	for _, element := range elements {
		if countElement(a.value, element) == 0 {
			a.chain.fail(AssertionFailure{
				Type:      AssertContainsElement,
				Actual:    &AssertionValue{a.value},
				Expected:  &AssertionValue{element},
				Reference: &AssertionValue{values},
				Errors: []error{
					errors.New("expected: array contains element from reference array"),
				},
			})
			return a
		}
	}

	for _, element := range a.value {
		if countElement(elements, element) == 0 {
			a.chain.fail(AssertionFailure{
				Type:      AssertNotContainsElement,
				Actual:    &AssertionValue{a.value},
				Expected:  &AssertionValue{element},
				Reference: &AssertionValue{values},
				Errors: []error{
					errors.New("expected: array does not contain elements" +
						" that are not present in reference array"),
				},
			})
			return a
		}
	}

	return a
}

// NotContainsOnly is opposite to ContainsOnly.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.NotContainsOnly(123)
//	array.NotContainsOnly(123, "foo", "bar")
//
// These calls are equivalent:
//
//	array.NotContainsOnly("a", "b")
//	array.NotContainsOnly("b", "a")
func (a *Array) NotContainsOnly(values ...interface{}) *Array {
	a.chain.enter("NotContainsOnly()")
	defer a.chain.leave()

	if a.chain.failed() {
		return a
	}

	elements, ok := canonArray(a.chain, values)
	if !ok {
		return a
	}

	different := false

	for _, element := range elements {
		if countElement(a.value, element) == 0 {
			different = true
			break
		}
	}

	for _, element := range a.value {
		if countElement(elements, element) == 0 {
			different = true
			break
		}
	}

	if !different {
		a.chain.fail(AssertionFailure{
			Type:      AssertNotEqual,
			Actual:    &AssertionValue{a.value},
			Expected:  &AssertionValue{values},
			Reference: &AssertionValue{values},
			Errors: []error{
				errors.New("expected:" +
					" array does not contain only elements from reference array" +
					" (at least one distinguishing element needed)"),
			},
		})
	}

	return a
}

// ContainsAny succeeds if array contains at least one element from the given elements.
// Before comparison, array and all elements are converted
// to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123, 123})
//	array.ContainsAny(123, "foo", "FOO") // success
//	array.ContainsAny("FOO") // failure
func (a *Array) ContainsAny(values ...interface{}) *Array {
	a.chain.enter("ContainsAny()")
	defer a.chain.leave()

	if a.chain.failed() {
		return a
	}

	elements, ok := canonArray(a.chain, values)
	if !ok {
		return a
	}

	foundAny := false

	for _, expected := range elements {
		if countElement(a.value, expected) > 0 {
			foundAny = true
			break
		}
	}

	if !foundAny {
		a.chain.fail(AssertionFailure{
			Type:      AssertContainsElement,
			Actual:    &AssertionValue{a.value},
			Reference: &AssertionValue{values},
			Errors: []error{
				errors.New("expected:" +
					" array contains at least one element from reference array"),
			},
		})
	}

	return a
}

// NotContainsAny succeeds if none of the given elements are in the array.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.NotContainsAny("bar", 124) // success
//	array.NotContainsAny(123) // failure
func (a *Array) NotContainsAny(values ...interface{}) *Array {
	a.chain.enter("NotContainsAny()")
	defer a.chain.leave()

	if a.chain.failed() {
		return a
	}

	elements, ok := canonArray(a.chain, values)
	if !ok {
		return a
	}

	for _, expected := range elements {
		if countElement(a.value, expected) > 0 {
			a.chain.fail(AssertionFailure{
				Type:      AssertNotContainsElement,
				Actual:    &AssertionValue{a.value},
				Expected:  &AssertionValue{expected},
				Reference: &AssertionValue{values},
				Errors: []error{
					errors.New("expected: array does not contain any elements from reference array"),
				},
			})
			return a
		}
	}

	return a
}

func countElement(array []interface{}, element interface{}) int {
	count := 0
	for _, e := range array {
		if reflect.DeepEqual(element, e) {
			count++
		}
	}
	return count
}
