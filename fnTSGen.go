package bluerpc

import (
	"fmt"
	"reflect"
	"strings"
)

func genTSFuncFromQuery(stringBuilder *strings.Builder, query, output interface{}, address string) {

	stringBuilder.WriteString("(")

	var dynamicSlugNames []string
	// for _, dynamicSlug := range dynamicSlugs {
	// 	dynamicSlugNames = append(dynamicSlugNames, dynamicSlug.Name)
	// }

	if !isInterpretedAsEmpty(query) {
		qpType := getType(query)
		stringBuilder.WriteString("query:")
		stringBuilder.WriteString(goToTsObj(qpType, dynamicSlugNames...))
		stringBuilder.WriteString(",")

	}
	stringBuilder.WriteString("headers?: HeadersInit,")
	stringBuilder.WriteString("):Promise<")

	generateFnOutputType(stringBuilder, output, dynamicSlugNames...)
	// address = addDynamicToAddress(address, QUERY, dynamicSlugs)
	generateQueryFnBody(stringBuilder, query != nil, address)
}

func genTSFuncFromMutation(stringBuilder *strings.Builder, query, input, output interface{}, address string) {

	stringBuilder.WriteString("(")

	var dynamicSlugNames []string
	// for _, dynamicSlug := range dynamicSlugs {
	// 	dynamicSlugNames = append(dynamicSlugNames, dynamicSlug.Name)
	// }

	isParams := !isInterpretedAsEmpty(query) || !isInterpretedAsEmpty(input)

	if isParams {
		stringBuilder.WriteString("parameters : {")
	}

	if !isInterpretedAsEmpty(query) {
		qpType := getType(query)
		if qpType.Kind() == reflect.Ptr {
			qpType = qpType.Elem()
		}

		stringBuilder.WriteString(fmt.Sprintf("query:%s,", goToTsObj(qpType, dynamicSlugNames...)))
	}
	if !isInterpretedAsEmpty(input) {
		inputType := getType(input)
		if inputType.Kind() == reflect.Ptr {
			inputType = inputType.Elem()
		}
		stringBuilder.WriteString(fmt.Sprintf("input:%s", goToTsObj(inputType, dynamicSlugNames...)))
	}

	if isParams {
		stringBuilder.WriteString("},")
	}
	stringBuilder.WriteString("headers?: HeadersInit,")

	stringBuilder.WriteString("):Promise<")
	generateFnOutputType(stringBuilder, output, dynamicSlugNames...)
	// address = addDynamicToAddress(address, MUTATION, dynamicSlugs)
	generateMutationFnBody(stringBuilder, isParams, address)
}
func generateFnOutputType(stringBuilder *strings.Builder, output any, dynamicSlugNames ...string) {
	if output != nil {
		outputType := getType(output)
		var tsType string
		if outputType.Kind() == reflect.Struct {
			tsType = goToTsObj(outputType, dynamicSlugNames...)
		} else {
			tsType = goTypeToTSType(outputType)
		}
		stringBuilder.WriteString(fmt.Sprintf("{body:%s,", tsType))
	} else {
		stringBuilder.WriteString("body:void,")
	}
	stringBuilder.WriteString("status: number, headers: Headers")
	stringBuilder.WriteString("}>=>")
}

// hasQuery here refers to if there's a query params variable placed
func generateQueryFnBody(stringBuilder *strings.Builder, hasQuery bool, address string) {

	stringBuilder.WriteString("{return rpcCall(")
	stringBuilder.WriteString("`" + address + "`")
	stringBuilder.WriteString(",'GET',")
	if hasQuery {
		stringBuilder.WriteString("{query}")
	} else {
		stringBuilder.WriteString("undefined")
	}
	stringBuilder.WriteString(",headers")
	stringBuilder.WriteString(")}")

}
func generateMutationFnBody(stringBuilder *strings.Builder, isParams bool, address string) {
	stringBuilder.WriteString("{return rpcCall(")
	stringBuilder.WriteString("`" + address + "`")
	stringBuilder.WriteString(",'POST',")
	if isParams {
		stringBuilder.WriteString("parameters")
	} else {
		stringBuilder.WriteString("undefined")
	}
	stringBuilder.WriteString(",headers")
	stringBuilder.WriteString(")}")
}

func getType(t interface{}) reflect.Type {
	if t == nil {
		return nil
	}
	typeOfT := reflect.TypeOf(t)

	if typeOfT == nil {
		return nil
	}
	// Check if the type is a pointer
	if typeOfT.Kind() == reflect.Ptr {
		// Get the type the pointer points to
		return typeOfT.Elem()
	}

	// Return the original type if it's not a pointer
	return typeOfT
}

// adds the needed dynamic typescript string to the address in the generated ts.
// func addDynamicToAddress(s string, method Method, slugInfos []dynamicSlugInfo) string {
// 	if len(slugInfos) == 0 {
// 		return s
// 	}
// 	splitStr := strings.Split(s, "/")

// 	for _, dsi := range slugInfos {
// 		var dynTsPart string
// 		switch method {
// 		case QUERY:
// 			dynTsPart = fmt.Sprintf(`query.%sSlug`, dsi.Name)
// 		case MUTATION:
// 			dynTsPart = fmt.Sprintf(`parameters.query.%sSlug`, dsi.Name)
// 		}
// 		splitStr[len(splitStr)-1-dsi.Position] = fmt.Sprintf(`${%s}`, dynTsPart)
// 	}
// 	return strings.Join(splitStr, "/")
// }

func isInterpretedAsEmpty(v interface{}) bool {
	// First, check if v is nil. This covers the case where v is nil itself.
	if v == nil {
		return true
	}

	// Use reflection to examine the type of v.
	val := reflect.ValueOf(v)
	typ := val.Type()

	// Check if v is a pointer.
	if typ.Kind() == reflect.Ptr {
		// Check if it's a nil pointer.
		if val.IsNil() {
			return true
		}

		// Check if the element type is an interface.
		elemType := val.Elem().Type()
		isPointerToInterface := elemType.Kind() == reflect.Interface
		return isPointerToInterface
	}

	// v is not a pointer and not nil.
	return false
}
