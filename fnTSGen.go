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
	}
	stringBuilder.WriteString("):Promise<")

	if output != nil {
		outputType := getType(output)
		stringBuilder.WriteString(goToTsObj(outputType, dynamicSlugNames...))
	} else {
		stringBuilder.WriteString("void")
	}
	stringBuilder.WriteString(">=>")
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
		stringBuilder.WriteString("}")
	}

	stringBuilder.WriteString("):Promise<")
	if output != nil {
		outputType := getType(output)
		if outputType.Kind() == reflect.Ptr {
			outputType = outputType.Elem()
		}
		stringBuilder.WriteString(goToTsObj(outputType))
	} else {
		stringBuilder.WriteString("void")
	}
	stringBuilder.WriteString(">=>")
	address = addDynamicToAddress(address, MUTATION, dynamicSlugs)
	generateMutationFnBody(stringBuilder, isParams, address)
}
func generateQueryFnBody(stringBuilder *strings.Builder, isQuery bool, address string) {

	stringBuilder.WriteString("{return rpcCall(")
	stringBuilder.WriteString("`" + address + "`")
	stringBuilder.WriteString(",")
	if isQuery {
		stringBuilder.WriteString("{query}")
	} else {
		stringBuilder.WriteString("undefined")
	}

	stringBuilder.WriteString(")}")

}
func generateMutationFnBody(stringBuilder *strings.Builder, isParams bool, address string) {
	stringBuilder.WriteString("{return rpcCall(")
	stringBuilder.WriteString("`" + address + "`")
	stringBuilder.WriteString(",")
	if isParams {
		stringBuilder.WriteString("parameters")
	} else {
		stringBuilder.WriteString("undefined")
	}

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
