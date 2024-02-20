package bluerpc

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// findIndex finds the index of a string in a slice. It returns -1 if the string is not found.
func findIndex(slice []string, val string) int {
	for i, item := range slice {
		if item == val {
			return i
		}
	}
	return -1
}

// splitStringOnSlash splits the string at each slash (unless it finds :/, meaning a dynamic route) and returns an array of substrings.
func splitStringOnSlash(s string) ([]string, error) {

	var result []string

	// Check if the string contains "/:" and find its position
	pos := strings.Index(s, "/:")
	if pos != -1 {
		// Split the string until the position of "/:"
		parts := strings.Split(s[:pos], "/")

		// Add non-empty parts to the result
		for _, part := range parts {
			if part != "" {
				result = append(result, "/"+part)
			}
		}

		// Add the remaining part of the string starting from "/:" as the last element
		result = append(result, s[pos:])
	} else {
		// Split the string at each slash if "/:" is not found
		parts := strings.Split(s, "/")

		// Add non-empty parts to the result
		for _, part := range parts {
			if part != "" {
				result = append(result, "/"+part)
			}
		}
	}

	return result, nil
}
func fillInField(field reflect.Value, key string, values ...any) error {

	fieldValue := reflect.ValueOf(values[0])
	if field.Type() == fieldValue.Type() {
		field.Set(fieldValue)
		return nil
	}
	fieldValueForcedToString := values[0].(string)

	fieldKind := field.Kind()
	switch fieldKind {
	case reflect.Slice:
		elemKind := field.Type().Elem().Kind()
		//this is disgusting but I have no clue how to go around the fact that I need to repeat the same thing multiple times
		// slice := reflect.MakeSlice(reflect.SliceOf(field.Type()), 0, len(values))

		if elemKind == reflect.Int {
			intSlice := make([]int, 0, len(values))

			for _, v := range values {
				intValue, err := strconv.Atoi(v.(string))
				if err != nil {
					return fmt.Errorf("invalid integer value '%s' for query parameter '%s'", v, key)
				}
				intSlice = append(intSlice, intValue)
			}
			field.Set(reflect.ValueOf(intSlice))
		}
	case reflect.String:
		field.SetString(fieldValueForcedToString)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bitSize, err := getByteSize(fieldKind)
		if err != nil {
			return err
		}
		intVal, err := strconv.ParseInt(fieldValueForcedToString, 10, bitSize)
		if err != nil {
			return fmt.Errorf("invalid integer value '%s' for query parameter '%s'", fieldValueForcedToString, key)
		}
		fmt.Println("setting int val", intVal)
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bitSize, err := getByteSize(fieldKind)
		if err != nil {
			return err
		}
		uIntVal, err := strconv.ParseUint(fieldValueForcedToString, 10, bitSize)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value '%s' for query parameter '%s'", fieldValueForcedToString, key)
		}
		field.SetUint(uIntVal)
	case reflect.Float32, reflect.Float64:
		bitSize, err := getByteSize(fieldKind)
		if err != nil {
			return err
		}
		floatVal, err := strconv.ParseFloat(fieldValueForcedToString, bitSize)
		if err != nil {
			return fmt.Errorf("invalid float value '%s' for query parameter '%s'", fieldValueForcedToString, key)
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(fieldValueForcedToString)
		if err != nil {
			return fmt.Errorf("invalid boolean value '%s' for query parameter '%s'", fieldValueForcedToString, key)
		}
		field.SetBool(boolVal)

	default:
		return fmt.Errorf("unsupported type '%s' for query parameter '%s'", field.Kind(), key)
	}
	return nil
}
func getByteSize(kind reflect.Kind) (int, error) {
	switch kind {
	case reflect.Int, reflect.Uint:
		return 0, nil
	case reflect.Int8, reflect.Uint8:
		return 8, nil
	case reflect.Int16, reflect.Uint16:
		return 16, nil
	case reflect.Int32, reflect.Uint32, reflect.Float32:
		return 32, nil
	case reflect.Int64, reflect.Uint64, reflect.Float64:
		return 64, nil
	}

	return 0, fmt.Errorf("passed kind is not a number, it is a %s", kind.String())
}
func findDynamicSlugs(s string) (info []dynamicSlugInfo) {

	routes := strings.Split(s, "/")

	//we start at 1 to avoid the first empty element. If the string starts with a slash the first element will be empty
	for i := 1; i < len(routes); i++ {
		route := routes[i]
		if route[0] == ':' {
			info = append(info, dynamicSlugInfo{
				Position: len(routes) - 1 - i,
				Name:     route[1:],
			})
		}

	}
	return info
}
func sliceStrContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
