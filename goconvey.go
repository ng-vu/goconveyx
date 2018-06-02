// Package goconveyx extends github.com/smartystreets/goconvey by providing a
// few more functions:
//
//     - ShouldDeepEqual
//     - ShouldResembleSlice
//     - ShouldResembleByKey
package goconveyx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-test/deep"
	"github.com/smartystreets/goconvey/convey"
)

// ShouldDeepEqual is the same as ShouldResemble with better error message.
func ShouldDeepEqual(actual interface{}, expected ...interface{}) string {
	res := convey.ShouldResemble(actual, expected...)
	if res == "" {
		return ""
	}

	const msg0 = "Expected: '%v'\nActual:   '%v'\n(Should deep equal)!"
	const msg1 = "Expected: '%v'\nActual:   '%v'\n(Should deep equal: %v)!"

	diff := deep.Equal(expected[0], actual)
	format := "Not match %d items: %v"
	if len(diff) == 0 {
		return fmt.Sprintf(msg0, spew.Sdump(expected[0]), spew.Sdump(actual)) + "\n" + res
	}
	if len(diff) == 1 {
		format = "Not match %d item: %v"
	}
	return fmt.Sprintf(msg1, spew.Sdump(expected[0]), spew.Sdump(actual),
		fmt.Sprintf(format, len(diff), spew.Sdump(diff)),
	)
}

// ShouldResembleSlice does deep equal comparison on two slices without ordering.
func ShouldResembleSlice(actual interface{}, expected ...interface{}) string {
	const msg = "Expected: '%v'\nActual:   '%v'\n(Should equal slice: %v)!"

	if len(expected) != 1 {
		return fmt.Sprintf("This assertion requires exactly %v comparison values (you provided %v).", 1, len(expected))
	}

	a := reflect.ValueOf(actual)
	e := reflect.ValueOf(expected[0])
	if a.Kind() != reflect.Slice || e.Kind() != reflect.Slice {
		return fmt.Sprintf(msg, spew.Sdump(expected[0]), spew.Sdump(actual),
			"Both must be slice")
	}

	if a.Len() != e.Len() {
		return fmt.Sprintf(msg, spew.Sdump(expected[0]), spew.Sdump(actual),
			"Length not equal")
	}

	count := 0
	indexes := make([]bool, a.Len())
	for i := 0; i < a.Len(); i++ {
		ai := a.Index(i)
		matched := false
		for j := 0; j < e.Len(); j++ {
			if ok := indexes[j]; ok {
				// already compared
				continue
			}
			ej := e.Index(j)
			if res := deep.Equal(ej.Interface(), ai.Interface()); res == nil {
				indexes[j] = true
				count++
				matched = true
				break
			}
		}
		if !matched {
			diff := deep.Equal(expected[0], actual)
			format := "Not match %d items: %v"
			if len(diff) == 1 {
				format = "Not match %d item: %v"
			}
			return fmt.Sprintf(msg, spew.Sdump(expected[0]), spew.Sdump(actual),
				fmt.Sprintf(format, len(diff), spew.Sdump(diff)),
			)
		}
	}
	if count != a.Len() {
		return fmt.Sprintf(msg, spew.Sdump(expected[0]), spew.Sdump(actual),
			"Slices are not equal")
	}
	return ""
}

// ShouldResembleByKey does deep equal comparison on two slices sorted by given
// key. It works on slices with map and struct as element. It's useful when you
// want to compare rows retrieved from database.
func ShouldResembleByKey(key string) func(actual interface{}, expected ...interface{}) string {
	const msg = "Expected: '%v'\nActual:   '%v'\n(Should equal slice: %v)!"

	return func(actual interface{}, expected ...interface{}) string {
		if len(expected) != 1 {
			return fmt.Sprintf("This assertion requires exactly %v comparison values (you provided %v).", 1, len(expected))
		}
		formatError := func(format string, args ...interface{}) string {
			errMsg := fmt.Sprintf(format, args...)
			return fmt.Sprintf(msg, spew.Sdump(expected[0]), spew.Sdump(actual), errMsg)
		}

		a := reflect.ValueOf(actual)
		e := reflect.ValueOf(expected[0])
		if a.Kind() != reflect.Slice || e.Kind() != reflect.Slice {
			return formatError("Both must be slice")
		}
		if errMsg := canGetKey(a.Type().Elem(), key); errMsg != "" {
			return formatError(errMsg)
		}
		if errMsg := canGetKey(e.Type().Elem(), key); errMsg != "" {
			return formatError(errMsg)
		}
		if a.Len() != e.Len() {
			return formatError("Length not equal")
		}

		collectIndexes := func(name string, list reflect.Value) (
			[]reflect.Value, map[interface{}]int, string,
		) {
			keys := make([]reflect.Value, list.Len())
			mapIndexes := make(map[interface{}]int)
			for i := 0; i < list.Len(); i++ {
				item := list.Index(i)
				switch item.Kind() {
				case reflect.Interface, reflect.Map, reflect.Ptr:
					if item.IsNil() {
						return nil, nil, formatError(
							"All items must not be nil (%v[%v] is nil)", name, i)
					}
				}

				itemKey := getKey(item, key)
				keys[i] = itemKey
				if !itemKey.IsValid() {
					return nil, nil, formatError(
						"Could not get key from %v[%v]", name, i)
				}

				keyValue := itemKey.Interface()
				if keyValue == nil {
					return nil, nil, formatError(
						"All item keys must not be nil (%v[%v].%v is nil)", name, i, key)
				}

				keyType := reflect.TypeOf(keyValue)
				if !keyType.Comparable() {
					return nil, nil, formatError(
						"All item keys must be comparable (%v[%v].%v is not, type is `%v`)",
						name, i, key, keyType)
				}
				if prev, ok := mapIndexes[keyValue]; ok {
					return nil, nil, formatError(
						"%v[%v] and %v[%v] has duplicated keys: `%v`",
						name, prev, name, i, keyValue)
				}
				mapIndexes[keyValue] = i
			}
			return keys, mapIndexes, ""
		}
		expectedKeys, _, errMsg := collectIndexes("expected", e)
		if errMsg != "" {
			return errMsg
		}
		_, mapActualIndexes, errMsg := collectIndexes("actual", a)
		if errMsg != "" {
			return errMsg
		}

		// Compare actual with the same order as expected
		for i, ekey := range expectedKeys {
			actualIndex, ok := mapActualIndexes[ekey.Interface()]
			if !ok {
				return formatError("Expected item with %v=`%v` but not found",
					key, ekey.Interface())
			}

			expectedItem := e.Index(i)
			actualItem := a.Index(actualIndex)

			diff := deep.Equal(actualItem.Interface(), expectedItem.Interface())
			if len(diff) > 0 {
				return formatError("Item with %v=`%v` is different: %v",
					key, ekey.Interface(), spew.Sdump(diff))
			}
		}
		return ""
	}
}

func canGetKey(t reflect.Type, key string) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Map, reflect.Interface:
		return ""
	case reflect.Struct:
		_, ok := t.FieldByName(key)
		if !ok {
			similar := ""
			for i, n := 0, t.NumField(); i < n; i++ {
				name := t.Field(i).Name
				if strings.ToLower(name) == strings.ToLower(key) {
					similar = name
				}
			}
			if similar != "" {
				return fmt.Sprintf(
					"Key `%v` not found in struct (but it has `%v`)",
					key, similar)
			}
			return fmt.Sprintf("Key `%v` not found in struct", key)
		}
		return ""
	}
	return "Both must be slice of struct, *struct, map or interface"
}

func getKey(v reflect.Value, key string) reflect.Value {
	if v.Kind() == reflect.Interface {
		v = reflect.ValueOf(v.Interface())
		if !v.IsValid() {
			return v
		}
	}
	v = reflect.Indirect(v)
	if !v.IsValid() {
		return v
	}

	switch v.Kind() {
	case reflect.Map:
		return v.MapIndex(reflect.ValueOf(key))
	case reflect.Struct:
		return v.FieldByName(key)
	default:
		return reflect.Value{}
	}
}
