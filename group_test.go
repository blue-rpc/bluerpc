package bluerpc

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
)

type group_test_local_query struct {
	Id                      string `paramName:"somethingotherthanId" validate:"required"`
	SomethingOtherThanQuery string `paramName:"query" validate:"required"`
}

type group_test_input struct {
	House string `paramName:"house" validate:"required"`
}

type group_test_output struct {
	FieldOneOut   string   `paramName:"fieldOneOut" validate:"required"`
	FieldTwoOut   string   `paramName:"fieldTwoOut" `
	FieldThreeOut string   `paramName:"fieldThreeOut" validate:"required"`
	FieldFourOut  []string `paramName:"fieldFourOut" `
}

func TestGroup(t *testing.T) {

	validate := validator.New(validator.WithRequiredStructEnabled())
	fmt.Println(DefaultColors.Green + "TESTING NESTED ROUTE" + DefaultColors.Reset)

	fmt.Println(DefaultColors.Green + "TESTING INVALID QUERY PARAMS" + DefaultColors.Reset)
	app := New(&Config{
		OutputPath:          "./some-file.ts",
		ValidatorFn:         validate.Struct,
		DisableInfoPrinting: true,
		DisableGenerateTS:   true,
	})

	proc := NewQuery[any, group_test_output](app, func(ctx *Ctx, query any) (*Res[group_test_output], error) {
		return &Res[group_test_output]{
			Status: 200,
			Header: Header{},
			Body: group_test_output{
				FieldOneOut:   "dwa",
				FieldTwoOut:   "dwadwa",
				FieldThreeOut: "dwadwadwa",
			},
		}, nil
	})
	depthOne := app.Router("/depth1")
	depthTwo := depthOne.Router("/depth2")

	proc.Attach(depthTwo, "/test")
	req, err := http.NewRequest("GET", "http://localhost:3000/depth1/depth2/test", nil)
	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not create a new request : %s", err.Error())
	}
	res, err := app.Test(req, ":3000")

	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not do the request : %s", err.Error())
	}

	if res.StatusCode > 300 {
		t.Fatalf(DefaultColors.Red+"Server did not respond with a 2xx status, actual status %d", res.StatusCode)

	}

	fmt.Println(DefaultColors.Green + "PASSED NESTED ROUTE TEST" + DefaultColors.Reset)

}
