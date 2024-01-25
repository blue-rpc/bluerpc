package bluerpc

import (
	"fmt"
	"net/http"
	"testing"
)

func TestUnauthorizedAuth(t *testing.T) {

	fmt.Println(DefaultColors.Green + "TESTING AUTHORIZATION" + DefaultColors.Reset)

	fmt.Println(DefaultColors.Green + "TESTING ACCESSING A PROTECTED ROUTE WITHOUT AUTHORIZATION " + DefaultColors.Reset)

	test_token := "test_token"
	app := New(&Config{
		OutputPath:          "./some-file.ts",
		DisableInfoPrinting: true,
		DisableGenerateTS:   true,
		Authorizer: NewAuth(func(ctx *Ctx) (any, error) {
			bearer := ctx.Get("Authorization")

			token := "Bearer " + test_token
			if bearer != token {
				return nil, fmt.Errorf("Unauthorized")
			}
			return User{
				Name: "hello",
			}, nil
		}),
	})

	protected := NewQuery[any, any](app, func(ctx *Ctx, query any) (*Res[any], error) {
		return &Res[any]{
			Status: 200,
			Header: Header{},
			Body: group_test_output{
				FieldOneOut:   "dwa",
				FieldTwoOut:   "dwadwa",
				FieldThreeOut: "dwadwadwa",
			},
		}, nil
	}).Protected()

	protected.Attach(app, "/protected")
	req, err := http.NewRequest("GET", "http://localhost:3000/protected", nil)
	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not create a new request : %s", err.Error())
	}
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not do the request : %s", err.Error())
	}

	if res.StatusCode != 401 {
		t.Fatalf(DefaultColors.Red+"Response status code for unauthorized access was NOT 401 : %s", err.Error())
	}

	fmt.Println(DefaultColors.Green + "PASSED ACCESSING A PROTECTED ROUTE WITHOUT AUTHORIZATION" + DefaultColors.Reset)

}
func TestAuthorizedAuth(t *testing.T) {

	fmt.Println(DefaultColors.Green + "TESTING AUTHORIZATION" + DefaultColors.Reset)

	fmt.Println(DefaultColors.Green + "TESTING ACCESSING A PROTECTED ROUTE WITHOUT AUTHORIZATION " + DefaultColors.Reset)

	test_token := "test_token"
	app := New(&Config{
		OutputPath:          "./some-file.ts",
		DisableInfoPrinting: true,
		DisableGenerateTS:   true,
		Authorizer: NewAuth(func(ctx *Ctx) (any, error) {
			bearer := ctx.Get("Authorization")

			token := "Bearer " + test_token
			if bearer != token {
				return nil, fmt.Errorf("Unauthorized")
			}
			return User{
				Name: "hello",
			}, nil
		}),
	})

	protected := NewQuery[any, any](app, func(ctx *Ctx, query any) (*Res[any], error) {
		return &Res[any]{
			Status: 200,
			Header: Header{},
			Body: group_test_output{
				FieldOneOut:   "dwa",
				FieldTwoOut:   "dwadwa",
				FieldThreeOut: "dwadwadwa",
			},
		}, nil
	}).Protected()

	protected.Attach(app, "/protected")
	req, err := http.NewRequest("GET", "http://localhost:3000/protected", nil)

	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not create a new request : %s", err.Error())
	}
	req.Header.Set("Authorization", "Bearer "+test_token)

	res, err := app.Test(req)
	if err != nil {
		t.Fatalf(DefaultColors.Red+"Could not do the request : %s", err.Error())
	}

	if res.StatusCode > 400 {
		t.Fatalf(DefaultColors.Red + "Response status code for authorized access was over 400 " + DefaultColors.Reset)
	}

	fmt.Println(DefaultColors.Green + "PASSED ACCESSING A PROTECTED ROUTE WITHOUT AUTHORIZATION" + DefaultColors.Reset)

}
