package bluerpc

import (
	"fmt"
	"reflect"
	"strings"
)

func genTSFuncFromQuery(stringBuilder *strings.Builder, query, output interface{}, address string) {

	stringBuilder.WriteString("(")
	if query != nil {
		qpType := getType(query)

		stringBuilder.WriteString("query:")
		stringBuilder.WriteString(GoFieldsToTSObj(qpType))
	}
	stringBuilder.WriteString("):Promise<")

	if output != nil {
		outputType := getType(output)
		stringBuilder.WriteString(GoFieldsToTSObj(outputType))
	} else {
		stringBuilder.WriteString("void")
	}
	stringBuilder.WriteString(">=>")
	generateQueryFnBody(stringBuilder, query != nil, address)
}

func genTSFuncFromMutation(stringBuilder *strings.Builder, query, input, output interface{}, address string) {

	stringBuilder.WriteString("(")

	isParams := query != nil || input != nil
	if isParams {
		stringBuilder.WriteString("parameters : {")
	}

	if query != nil {
		qpType := getType(query)
		if qpType.Kind() == reflect.Ptr {
			qpType = qpType.Elem()
		}
		stringBuilder.WriteString(fmt.Sprintf("query:%s,", GoFieldsToTSObj(qpType)))
	}
	if input != nil {
		inputType := getType(input)
		if inputType.Kind() == reflect.Ptr {
			inputType = inputType.Elem()
		}
		stringBuilder.WriteString(fmt.Sprintf("input:%s", GoFieldsToTSObj(inputType)))
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
		stringBuilder.WriteString(GoFieldsToTSObj(outputType))
	} else {
		stringBuilder.WriteString("void")
	}
	stringBuilder.WriteString(">=>")
	generateMutationFnBody(stringBuilder, isParams, address)
}
func generateQueryFnBody(stringBuilder *strings.Builder, isQuery bool, address string) {
	stringBuilder.WriteString("{return rpcCall(")
	if isQuery {
		stringBuilder.WriteString("{query}")
	} else {
		stringBuilder.WriteString("undefined")
	}
	stringBuilder.WriteString(",")
	stringBuilder.WriteString("'" + address + `'`)

	stringBuilder.WriteString(")}")

}
func generateMutationFnBody(stringBuilder *strings.Builder, isParams bool, address string) {
	stringBuilder.WriteString("{return rpcCall(")
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
