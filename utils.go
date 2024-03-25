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

// splitStringOnSlash splits the string at each slash (unless it finds /{...}, meaning a dynamic route) and returns an array of substrings.
func splitStringOnSlash(s string) ([]string, error) {

	var result []string
	//check if the string contains /{}
	pos := strings.Index(s, "/{")
	if pos != -1 {
		// Split the string until the position of "/{"
		parts := strings.Split(s[:pos], "/")

		// Add non-empty parts to the result
		for _, part := range parts {
			if part != "" {
				result = append(result, "/"+part)
			}
		}

		// Add the remaining part of the string starting from "/{" as the last element
		result = append(result, s[pos:])
		return result, nil
	}

	parts := strings.Split(s, "/")

	// Add non-empty parts to the result
	for _, part := range parts {
		if part != "" {
			result = append(result, "/"+part)
		}
	}

	return result, nil
}
func setField(field reflect.Value, values []string) error {
	if len(values) == 0 {
		return nil
	}
	kind := field.Kind()
	switch field.Kind() {
	case reflect.Slice, reflect.Array:
		elemType := field.Type().Elem()
		collection := reflect.MakeSlice(reflect.SliceOf(elemType), len(values), len(values))
		for i, value := range values {
			elem := reflect.New(elemType).Elem()
			if err := setField(elem, []string{value}); err != nil {
				return fmt.Errorf("failed to set slice/array element: %v", err)
			}
			collection.Index(i).Set(elem)
		}
		if field.Kind() == reflect.Slice {
			field.Set(collection)
		} else if field.Kind() == reflect.Array && field.Len() == collection.Len() {
			reflect.Copy(field, collection)
		}
	case reflect.Map:
		// For simplicity, assuming string keys and handling only string values here
		mapType := field.Type()
		keyType := mapType.Key()
		elemType := mapType.Elem()
		newMap := reflect.MakeMap(mapType)
		for _, value := range values {
			key := reflect.ValueOf(value) // Simplified: using value as key
			elem := reflect.New(elemType).Elem()
			if err := setField(elem, []string{value}); err != nil {
				return fmt.Errorf("failed to set map element: %v", err)
			}
			newMap.SetMapIndex(key.Convert(keyType), elem)
		}
		field.Set(newMap)
	case reflect.Struct:
		// For simplicity, assigning values to fields by their index
		for i := 0; i < field.NumField(); i++ {
			if len(values) > i {
				if err := setField(field.Field(i), []string{values[i]}); err != nil {
					return fmt.Errorf("failed to set struct field: %v", err)
				}
			}
		}
	case reflect.String:
		field.SetString(values[0])
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		byteSize, err := getByteSize(kind)
		if err != nil {
			return err
		}
		intValue, err := strconv.ParseInt(values[0], 10, byteSize)
		if err != nil {
			return err
		}
		field.SetInt(intValue)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		byteSize, err := getByteSize(kind)
		if err != nil {
			return err
		}
		uintValue, err := strconv.ParseUint(values[0], 10, byteSize)
		if err != nil {
			return err
		}
		field.SetUint(uintValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(values[0])
		if err != nil {
			field.SetBool(boolValue)
		} else {
			return err
		}
	case reflect.Float32, reflect.Float64:
		byteSize, err := getByteSize(kind)
		if err != nil {
			return err
		}
		floatValue, err := strconv.ParseFloat(values[0], byteSize)
		if err != nil {
			return err
		}
		field.SetFloat(floatValue)
	default:
		return fmt.Errorf("unsupported field type %s", field.Type())
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

func sliceStrContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
