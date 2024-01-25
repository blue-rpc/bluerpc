package bluerpc

import (
	"fmt"
	"reflect"
	"strings"
)

func genTSFuncFromQuery(stringBuilder *strings.Builder, query, output interface{}, address string, dynamicSlugs []dynamicSlugInfo) {

	stringBuilder.WriteString("(")

	var dynamicSlugNames []string
	for _, dynamicSlug := range dynamicSlugs {
		dynamicSlugNames = append(dynamicSlugNames, dynamicSlug.Name)
	}

	if query != nil {
		qpType := getType(query)
		stringBuilder.WriteString("query:")
		stringBuilder.WriteString(goToTsObj(qpType, dynamicSlugNames...))
		stringBuilder.WriteString(",")

	}
	stringBuilder.WriteString("headers?: HeadersInit,")
	stringBuilder.WriteString("):Promise<")

	generateFnOutputType(stringBuilder, output, dynamicSlugNames...)
	address = addDynamicToAddress(address, QUERY, dynamicSlugs)
	generateQueryFnBody(stringBuilder, query != nil, address)
}

func genTSFuncFromMutation(stringBuilder *strings.Builder, query, input, output interface{}, address string, dynamicSlugs []dynamicSlugInfo) {

	stringBuilder.WriteString("(")

	var dynamicSlugNames []string
	for _, dynamicSlug := range dynamicSlugs {
		dynamicSlugNames = append(dynamicSlugNames, dynamicSlug.Name)
	}

	isParams := query != nil || input != nil

	fmt.Println("is params from fnTsGen", isParams, "query", query, "input", input)
	if isParams {
		stringBuilder.WriteString("parameters : {")
	}

	if query != nil {
		qpType := getType(query)
		if qpType.Kind() == reflect.Ptr {
			qpType = qpType.Elem()
		}

		stringBuilder.WriteString(fmt.Sprintf("query:%s,", goToTsObj(qpType, dynamicSlugNames...)))
	}
	if input != nil {
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
	address = addDynamicToAddress(address, MUTATION, dynamicSlugs)
	generateMutationFnBody(stringBuilder, isParams, address)
}
func generateFnOutputType(stringBuilder *strings.Builder, output any, dynamicSlugNames ...string) {
	if output != nil {
		outputType := getType(output)
		stringBuilder.WriteString(fmt.Sprintf("{body:%s,", goToTsObj(outputType, dynamicSlugNames...)))
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
func addDynamicToAddress(s string, method Method, slugInfos []dynamicSlugInfo) string {
	if len(slugInfos) == 0 {
		return s
	}
	splitStr := strings.Split(s, "/")

	for _, dsi := range slugInfos {
		var dynTsPart string
		switch method {
		case QUERY:
			dynTsPart = fmt.Sprintf(`query.%sSlug`, dsi.Name)
		case MUTATION:
			dynTsPart = fmt.Sprintf(`parameters.query.%sSlug`, dsi.Name)
		}
		splitStr[len(splitStr)-1-dsi.Position] = fmt.Sprintf(`${%s}`, dynTsPart)
	}
	return strings.Join(splitStr, "/")
}
