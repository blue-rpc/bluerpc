package bluerpc

import (
	"reflect"
)

type Method string

var (
	QUERY    Method = "query"
	MUTATION Method = "mutation"
)

type Procedure[queryParams any, input any, output any] struct {
	queryParamsSchema *queryParams
	inputSchema       *input
	outputSchema      *output

	method      Method
	app         *App
	validatorFn *validatorFn

	queryHandler    Query[queryParams, output]
	mutationHandler Mutation[queryParams, input, output]
}
type ProcedureInfo struct {
	method      Method
	validatorFn *validatorFn

	querySchema  interface{}
	inputSchema  interface{}
	outputSchema interface{}
	handler      func(ctx *Ctx) error
}

// Creates a new query procedure that can be attached to groups / app root.
// The generic arguments specify the structure for validating query parameters (the query Params, the body of the request and the resulting handler output).
// Use any to avoid validation
func NewMutation[queryParams any, input any, output any](app *App, mutation Mutation[queryParams, input, output]) *Procedure[queryParams, input, output] {

	var queryParamsInstance *queryParams
	var inputInstance *input
	var outputInstance *output

	if getType(new(queryParams)).Kind() == reflect.Struct {
		temp := new(queryParams)
		queryParamsInstance = temp
	}

	if getType(new(input)).Kind() == reflect.Struct {
		temp := new(input)
		inputInstance = temp
	}

	// Check if output is a struct
	if getType(new(output)).Kind() == reflect.Struct {
		temp := new(output)
		outputInstance = temp
	}

	return &Procedure[queryParams, input, output]{
		app:               app,
		validatorFn:       &app.config.ValidatorFn,
		inputSchema:       inputInstance,
		queryParamsSchema: queryParamsInstance,
		outputSchema:      outputInstance,
		method:            MUTATION,
		mutationHandler:   mutation,
	}
}

// Creates a new query procedure that can be attached to groups / app root.
// The generic arguments specify the structure for validating query parameters (the query Params and the resulting handler output).
// Use any to avoid validation
func NewQuery[queryParams any, output any](app *App, query Query[queryParams, output]) *Procedure[queryParams, any, output] {

	queryParamsInstance := new(queryParams)
	outputInstance := new(output)

	return &Procedure[queryParams, any, output]{
		app:               app,
		validatorFn:       &app.config.ValidatorFn,
		outputSchema:      outputInstance,
		queryParamsSchema: queryParamsInstance,
		method:            QUERY,
		queryHandler:      query,
	}
}
