/*
Small library on top of reflect for make lookups to Structs or Maps. Using a
very simple DSL you can access to any property, key or value of any value of Go.
*/
package lookup

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

const (
	SplitToken     = "."
	IndexCloseChar = "]"
	IndexOpenChar  = "["
)

var (
	ErrMalformedIndex    = errors.New("Malformed index key")
	ErrInvalidIndexUsage = errors.New("Invalid index key usage")
	ErrKeyNotFound       = errors.New("Unable to find the key")
)

// LookupString performs a lookup into a value, using a string. Same as `Loookup`
// but using a string with the keys separated by `.`
func LookupString(i interface{}, path string) (reflect.Value, error) {
	return Lookup(i, strings.Split(path, SplitToken)...)
}

// Lookup performs a lookup into a value, using a path of keys. The key should
// match with a Field or a MapIndex. For slice you can use the syntax key[index]
// to access a specific index. If one key owns to a slice and an index is not
// specificied the rest of the path will be apllied to evaley value of the
// slice, and the value will be merged into a slice.
func Lookup(i interface{}, path ...string) (reflect.Value, error) {
	value := reflect.ValueOf(i)
	var parent reflect.Value
	var err error

	for i, part := range path {
		parent = value

		value, err = getValueByName(value, part)
		if err == nil {
			continue
		}

		if !isAggregable(parent) {
			break
		}

		value, err = aggreateAggregableValue(parent, path[i:])

		break
	}

	return value, err
}

func getValueByName(v reflect.Value, key string) (reflect.Value, error) {
	var value reflect.Value
	var index int
	var err error

	key, index, err = parseIndex(key)
	if err != nil {
		return value, err
	}
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return getValueByName(v.Elem(), key)
	case reflect.Struct:
		value = v.FieldByName(key)
	case reflect.Map:
		kValue := reflect.Indirect(reflect.New(v.Type().Key()))
		kValue.SetString(key)
		value = v.MapIndex(kValue)
	}

	if !value.IsValid() {
		return reflect.Value{}, ErrKeyNotFound
	}

	if index != -1 {
		if value.Type().Kind() != reflect.Slice {
			return reflect.Value{}, ErrInvalidIndexUsage
		}

		value = value.Index(index)
	}

	if value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
		value = value.Elem()
	}

	return value, nil
}

func aggreateAggregableValue(v reflect.Value, path []string) (reflect.Value, error) {

	values := make([]reflect.Value, 0)

	l := v.Len()

	for i := 0; i < l; i++ {
		value, err := Lookup(v.Index(i).Interface(), path...)

		if err != nil {
			return reflect.Value{}, err
		}

		values = append(values, value)
	}

	return mergeValue(values), nil
}

func mergeValue(values []reflect.Value) reflect.Value {
	values = removeZeroValues(values)
	l := len(values)
	if l == 0 {
		return reflect.Value{}
	}

	sample := values[0]
	mergeable := isMergeable(sample)

	t := sample.Type()
	if mergeable {
		t = t.Elem()
	}

	value := reflect.MakeSlice(reflect.SliceOf(t), 0, 0)
	for i := 0; i < l; i++ {
		if !values[i].IsValid() {
			continue
		}

		if mergeable {
			value = reflect.AppendSlice(value, values[i])
		} else {
			value = reflect.Append(value, values[i])
		}
	}

	return value
}

func removeZeroValues(values []reflect.Value) []reflect.Value {
	l := len(values)

	var v []reflect.Value
	for i := 0; i < l; i++ {
		if values[i].IsValid() {
			v = append(v, values[i])
		}
	}

	return v
}

func isAggregable(v reflect.Value) bool {
	k := v.Kind()

	return k == reflect.Struct || k == reflect.Map || k == reflect.Slice
}

func isMergeable(v reflect.Value) bool {
	k := v.Kind()
	return k == reflect.Map || k == reflect.Slice
}

func hasIndex(s string) bool {
	return strings.Index(s, IndexOpenChar) != -1
}

func parseIndex(s string) (string, int, error) {
	start := strings.Index(s, IndexOpenChar)
	end := strings.Index(s, IndexCloseChar)

	if start == -1 && end == -1 {
		return s, -1, nil
	}

	if (start != -1 && end == -1) || (start == -1 && end != -1) {
		return "", -1, ErrMalformedIndex
	}

	index, err := strconv.Atoi(s[start+1 : end])
	if err != nil {
		return "", -1, ErrMalformedIndex
	}

	return s[:start], index, nil
}
