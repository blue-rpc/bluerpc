package bluerpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
)

type test_query struct {
	Something string `paramName:"query" validate:"required"`
}
type procedure_test_input struct {
	House string `paramName:"House" validate:"required"`
}
type procedure_test_output struct {
	FieldOneOut   string   `paramName:"fieldOneOut" validate:"required"`
	FieldTwoOut   string   `paramName:"fieldTwoOut" `
	FieldThreeOut string   `paramName:"fieldThreeOut" validate:"required"`
	FieldFourOut  []string `paramName:"fieldFourOut" `
}

func TestQuery(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	fmt.Println(DefaultColors.Green + "TESTING INVALID QUERY PARAMS")
	app := New(&Config{
		OutputPath:  "./some-file.ts",
		ValidatorFn: validate.Struct,
	})

	proc := NewQuery[test_query, procedure_test_output](app, func(ctx *Ctx, query test_query) (*Res[procedure_test_output], error) {
		return &Res[procedure_test_output]{
			Status: 200,
			Header: Header{},
			Body: procedure_test_output{
				FieldOneOut:   "dwa",
				FieldTwoOut:   "dwadwa",
				FieldThreeOut: "dwadwadwa",
			},
		}, nil
	})
	proc.Attach(app, "/test")

	// app.Listen(":3000")
	req, err := http.NewRequest("GET", "http://localhost:8080/test", nil)
	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not create a new request", err.Error())
	}
	res, err := app.Test(req)

	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not do the request", err.Error())
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not read the body", err.Error())
	}
	fmt.Println("body", string(body))
	type DefaultResError struct {
		Message string `json:"message"`
	}
	var resError DefaultResError
	if err := json.Unmarshal(body, &resError); err != nil {
		t.Fatalf(DefaultColors.Red+"Failed to unmarshal response: %v", err)
	}
	if resError.Message == "" {
		t.Fatalf(DefaultColors.Red + "The server responded without an error")
	}

	fmt.Println(DefaultColors.Green + "PASSED INVALID QUERY")

	// TESTING VALID QUERY PARAMS
	fmt.Println(DefaultColors.Green + "TESTING VALID QUERY PARAMS")
	req, err = http.NewRequest("GET", "http://localhost:8080/test?query=dwa", nil)
	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not create a new request", err.Error())
	}
	res, err = app.Test(req)
	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not do the request", err.Error())
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not read the body", err.Error())
	}

	var output procedure_test_output
	if err := json.Unmarshal(body, &output); err != nil {
		t.Fatalf(DefaultColors.Red+"Failed to unmarshal response: %v", err)
	}

	if output.FieldOneOut == "" || output.FieldTwoOut == "" || output.FieldThreeOut == "" {
		t.Fatalf(DefaultColors.Red+"The server responded with an invalid output response", string(body))
	}

	fmt.Println(DefaultColors.Green + "PASSED VALID QUERY")
}

func TestMutation(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	fmt.Println(DefaultColors.Green + "TESTING INVALID MUTATION PARAMS" + DefaultColors.Reset)
	app := New(&Config{
		OutputPath:  "./some-file.ts",
		ValidatorFn: validate.Struct,
	})

	proc := NewMutation[test_query, procedure_test_input, procedure_test_output](app, func(ctx *Ctx, query test_query, input procedure_test_input) (*Res[procedure_test_output], error) {

		return &Res[procedure_test_output]{
			Status: 200,
			Body: procedure_test_output{
				FieldOneOut:   "dwaawdwa",
				FieldTwoOut:   "dwa",
				FieldThreeOut: "dawdwadwadwa",
			},
		}, nil

	})
	proc.Attach(app, "/test")
	// app.Listen(":3000")

	inputData := procedure_test_input{
		House: "hello world",
	}

	jsonData, err := json.Marshal(inputData)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	res, err := app.Test(req)

	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not do the request", err.Error())
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not read the body", err.Error())
	}
	type DefaultResError struct {
		Message string
	}
	var resError DefaultResError
	fmt.Println("body for mutation", string(body))
	if err := json.Unmarshal(body, &resError); err != nil {
		t.Fatalf(DefaultColors.Red+"Failed to unmarshal response: %v", err)
	}
	if resError.Message == "" {
		t.Fatalf(DefaultColors.Red+"The server responded without an error, %s", string(body))
	}

	fmt.Println(DefaultColors.Green + "PASSED INVALID MUTATION")

	// TESTING VALID QUERY PARAMS

	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}
	fmt.Println(DefaultColors.Green + "TESTING VALID MUTATION PARAMS" + DefaultColors.Reset)
	req, err = http.NewRequest("POST", "http://localhost:8080/test?query=dwa", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not create a new request", err.Error())
	}
	res, err = app.Test(req)
	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not do the request", err.Error())
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not read the body", err.Error())
	}

	var output procedure_test_output
	if err := json.Unmarshal(body, &output); err != nil {
		t.Fatalf(DefaultColors.Red+"Failed to unmarshal response: %v", err)
	}

	if output.FieldOneOut == "" || output.FieldTwoOut == "" || output.FieldThreeOut == "" {
		t.Fatalf(DefaultColors.Red+"The server responded with an invalid output response, %s", string(body))
	}

	fmt.Println(DefaultColors.Green + "PASSED VALID MUTATION")
	fmt.Println(DefaultColors.Green + "TESTING INVALID OUTPUT")

	fakeProc := NewMutation[test_query, procedure_test_input, procedure_test_output](app, func(ctx *Ctx, query test_query, input procedure_test_input) (*Res[procedure_test_output], error) {

		return &Res[procedure_test_output]{
			Status: 200,
			Body: procedure_test_output{
				FieldOneOut:   "",
				FieldTwoOut:   "dwa",
				FieldThreeOut: query.Something,
			},
		}, nil

	})
	fakeProc.Attach(app, "/error")

	req, err = http.NewRequest("POST", "http://localhost:8080/error?query=dwa", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not create a new request", err.Error())
	}
	res, err = app.Test(req)
	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not do the request", err.Error())
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not read the body", err.Error())
	}

	var err_output DefaultResError
	fmt.Println("body", string(body))
	if err := json.Unmarshal(body, &err_output); err != nil {
		t.Fatalf(DefaultColors.Red+"Failed to unmarshal response: %v", err)
	}

	if err_output.Message == "" {
		t.Fatalf(DefaultColors.Red+"The body output error response is not proper : %v", err)

	}

	fmt.Println(DefaultColors.Green + "PASSED INVALID OUTPUT" + DefaultColors.Reset)

}
