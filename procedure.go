package bluerpc

import (
	"fmt"
	"reflect"
)

type Method string

var (
	QUERY    Method = "query"
	MUTATION Method = "mutation"
)

type Procedure[query any, input any, output any] struct {
	hasQuery  bool
	hasInput  bool
	hasOutput bool

	method      Method
	app         *App
	validatorFn *validatorFn

	acceptedContentType []string
	queryHandler        Query[query, output]
	mutationHandler     Mutation[query, input, output]

	authorizer *Authorizer
	protected  bool
}

// position refers to the position of the slug FROM THE END
// meaning in /api/:dynamic ":dynamic" will be in POSITION 0
type dynamicSlugInfo struct {
	Position int
	Name     string
}
type ProcedureInfo struct {
	method      Method
	validatorFn *validatorFn

	querySchema  interface{}
	inputSchema  interface{}
	outputSchema interface{}

	dynamicSlugs []dynamicSlugInfo
	handler      func(ctx *Ctx) error
	protected    bool
	authorizer   *Authorizer
}

// Creates a new query procedure that can be attached to groups / app root.
// The generic arguments specify the structure for validating query parameters (the query Params, the body of the request and the resulting handler output).
// Use any to avoid validation
func NewMutation[query any, input any, output any](app *App, mutation Mutation[query, input, output]) *Procedure[query, input, output] {

	var queryInstance query
	checkIfQueryStruct(queryInstance)

	return &Procedure[query, input, output]{
		app:                 app,
		validatorFn:         &app.config.ValidatorFn,
		method:              MUTATION,
		mutationHandler:     mutation,
		acceptedContentType: []string{ApplicationJSON, ApplicationForm},
		hasQuery:            !isEmptyInterface[query](),
		hasInput:            !isEmptyInterface[input](),
		hasOutput:           !isEmptyInterface[output](),
	}
}

// Creates a new query procedure that can be attached to groups / app root.
// The generic arguments specify the structure for validating query parameters (the query Params and the resulting handler output).
// Use any to avoid validation
func NewQuery[query any, output any](app *App, queryFn Query[query, output]) *Procedure[query, any, output] {

	var queryInstance query
	checkIfQueryStruct(queryInstance)

	return &Procedure[query, any, output]{
		app:                 app,
		validatorFn:         &app.config.ValidatorFn,
		method:              QUERY,
		queryHandler:        queryFn,
		acceptedContentType: []string{ApplicationJSON, ApplicationForm},
		hasQuery:            !isEmptyInterface[query](),
		hasOutput:           !isEmptyInterface[output](),
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

func isEmptyInterface[T any]() bool {
	// Create a zero value of type T
	var zeroT T

	// Use reflect to get the type of zeroT
	t := reflect.TypeOf(zeroT)

	// Return true if t is an interface and has no methods
	return t == nil || (t.Kind() == reflect.Interface && t.NumMethod() == 0)
}

// This function should panic if the query params are not a struct or of type any (interface{})
func checkIfQueryStruct[query any](arg query) {
	queryT := reflect.TypeOf(arg) // Reflect on the zero value, not T directly

	//if this is an empty interface return
	if queryT == nil || (queryT.Kind() == reflect.Interface && queryT.NumMethod() == 0) {
		return
	}

	if queryT.Kind() == reflect.Ptr && queryT.Elem().Kind() != reflect.Struct {
		panic(fmt.Sprintf("generic argument Query must be a struct or a pointer to a struct or any (interface{}), got %s", queryT.Elem().Kind()))
	} else if queryT.Kind() != reflect.Struct {
		panic(fmt.Sprintf("generic argument Query must be a struct, got %s", queryT.Kind()))
	}
}

// Turns the procedure into a protected procedure, meaning your authorization handler will run before this runs
func (p *Procedure[query, input, output]) Protected() *Procedure[query, input, output] {
	p.protected = true
	return p
}
func (p *Procedure[query, input, output]) Authorizer(a *Authorizer) *Procedure[query, input, output] {
	p.authorizer = a
	return p
}
