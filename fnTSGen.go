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
		fmt.Println("dynamic slug name", dynamicSlug.Name)
		fmt.Println("dynamic slug pos", dynamicSlug.Position)

		dynamicSlugNames = append(dynamicSlugNames, dynamicSlug.Name)
	}

	if query != nil {
		qpType := getType(query)
		stringBuilder.WriteString("query:")
		stringBuilder.WriteString(goToTsField(qpType, dynamicSlugNames...))
	}
	stringBuilder.WriteString("):Promise<")

	if output != nil {
		outputType := getType(output)
		stringBuilder.WriteString(goToTsField(outputType, dynamicSlugNames...))
	} else {
		stringBuilder.WriteString("void")
	}
	stringBuilder.WriteString(">=>")
	address = addDynamicToAddress(address, dynamicSlugs)
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
		stringBuilder.WriteString(fmt.Sprintf("query:%s,", goToTsField(qpType, dynamicSlugNames...)))
	}
	if input != nil {
		inputType := getType(input)
		if inputType.Kind() == reflect.Ptr {
			inputType = inputType.Elem()
		}
		stringBuilder.WriteString(fmt.Sprintf("input:%s", goToTsField(inputType, dynamicSlugNames...)))
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
		stringBuilder.WriteString(goToTsField(outputType))
	} else {
		stringBuilder.WriteString("void")
	}
	stringBuilder.WriteString(">=>")
	address = addDynamicToAddress(address, dynamicSlugs)
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
	stringBuilder.WriteString(",")
	stringBuilder.WriteString("'" + address + `'`)

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

func addDynamicToAddress(s string, slugInfos []dynamicSlugInfo) string {
	if len(slugInfos) == 0 {
		return s
	}
	splitStr := strings.Split(s, "/")

	for _, dsi := range slugInfos {
		splitStr[len(splitStr)-1-dsi.Position] = fmt.Sprintf(`${query.%sSlug}`, dsi.Name)
	}
	return strings.Join(splitStr, "/")
}
