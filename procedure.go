package bluerpc

import (
	"reflect"
)

type Method string

var (
	QUERY    Method = "query"
	MUTATION Method = "mutation"
)

type Procedure[query any, input any, output any] struct {
	querySchema  *query
	inputSchema  *input
	outputSchema *output

	method      Method
	app         *App
	validatorFn *validatorFn

	acceptedContentType []string
	queryHandler        Query[query, output]
	mutationHandler     Mutation[query, input, output]
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
func NewMutation[query any, input any, output any](app *App, mutation Mutation[query, input, output]) *Procedure[query, input, output] {

	var queryInstance *query
	var inputInstance *input
	var outputInstance *output

	if getType(new(query)).Kind() == reflect.Struct {
		temp := new(query)
		queryInstance = temp
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

	return &Procedure[query, input, output]{
		app:                 app,
		validatorFn:         &app.config.ValidatorFn,
		inputSchema:         inputInstance,
		querySchema:         queryInstance,
		outputSchema:        outputInstance,
		method:              MUTATION,
		mutationHandler:     mutation,
		acceptedContentType: []string{ApplicationJSON, ApplicationForm},
	}
}

// Creates a new query procedure that can be attached to groups / app root.
// The generic arguments specify the structure for validating query parameters (the query Params and the resulting handler output).
// Use any to avoid validation
func NewQuery[query any, output any](app *App, queryFn Query[query, output]) *Procedure[query, any, output] {

	queryInstance := new(query)
	outputInstance := new(output)

	return &Procedure[query, any, output]{
		app:                 app,
		validatorFn:         &app.config.ValidatorFn,
		outputSchema:        outputInstance,
		querySchema:         queryInstance,
		method:              QUERY,
		queryHandler:        queryFn,
		acceptedContentType: []string{ApplicationJSON, ApplicationForm},
	}
}

// Changes the validator function for this particular procedure
func (p *Procedure[query, input, output]) Validator(fn validatorFn) {
	p.validatorFn = &fn
}

// determines which content type should a request have in order for it to be valid.
func (p *Procedure[query, input, output]) AcceptedContentType(contentTypes ...string) {
	p.acceptedContentType = contentTypes
}
