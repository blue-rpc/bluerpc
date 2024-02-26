package bluerpc

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
)

func goToTsObj(someStruct reflect.Type, dynamicSlugNames ...string) string {
	stringBuilder := strings.Builder{}

	if someStruct == nil {
		return "any"
	}
	if someStruct.Kind() == reflect.Ptr {
		someStruct = someStruct.Elem()
	}
	if someStruct.Kind() == reflect.Interface {
		return "any"
	} else if someStruct.Kind() != reflect.Struct {
		log.Panicf(" i though this was supposed to be a struct, it is %s", someStruct.Kind())
	}

	stringBuilder.WriteString("{")

	for i := 0; i < someStruct.NumField(); i++ {
		field := someStruct.Field(i)
		fieldName := field.Name
		fieldType := field.Type

		paramName := field.Tag.Get("paramName")

		if paramName != "" {
			regex := regexp.MustCompile("[^a-zA-Z]+")
			fieldName = regex.ReplaceAllString(paramName, "")
		}
		_, hasRequired := field.Tag.Lookup("required")
		validateTag := field.Tag.Get("validate")
		hasValidateRequired := strings.Contains(validateTag, "required")

		// If the name is from a dynamic slug then just put Slug at the end
		//Because dynamic slugs params are always required I put this else if here
		if len(dynamicSlugNames) > 0 && sliceStrContains(dynamicSlugNames, fieldName) {
			fieldName += "Slug"
		} else if !hasRequired && !hasValidateRequired {
			fieldName += "?"
		}

		// Append TypeScript field definition to the StringBuilder
		stringBuilder.WriteString(fmt.Sprintf(" %s: %s", fieldName, goTypeToTSType(fieldType)))

		stringBuilder.WriteString(",")

	}
	stringBuilder.WriteString("}")
	return stringBuilder.String()
}

func goTypeToTSType(t reflect.Type) string {
	if t == nil {
		return "any"
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// Check if the type is a named type and if its Kind is one of the basic types
	if t.Name() != "" && t.Kind() != reflect.Struct && t.Kind() != reflect.Interface {
		switch t.Kind() {
		case reflect.String:
			return "string"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
			return "number"
		case reflect.Bool:
			return "boolean"
		}
	}

	switch t.Kind() {

	case reflect.Slice, reflect.Array:
		elemType := goTypeToTSType(t.Elem())
		return fmt.Sprintf("Array<%s>", elemType)
	case reflect.Map:
		keyType := goTypeToTSType(t.Key())
		valueType := goTypeToTSType(t.Elem())
		return fmt.Sprintf("Record<%s, %s>", keyType, valueType)
	case reflect.Struct:
		return goToTsObj(t)
	case reflect.Interface:
		return "any"
	default:
		return "any"
	}
}
