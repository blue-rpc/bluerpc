package bluerpc

import (
	"fmt"
	"net/http"
)

type Route interface {
	getAbsPath() string
	addProcedure(string, *ProcedureInfo)
	getValidatorFn() *validatorFn
	getApp() *App
}

func (proc *Procedure[query, input, output]) Attach(route Route, slug string) {

	absPath := route.getAbsPath()
	validatorFn := *proc.validatorFn
	fullRoute := absPath + slug
	fullHandler := func(c *Ctx) error {
		queryParams, err := validateQuery(c, proc, slug)
		if err != nil {
			return err
		}
		var res *Res[output]

		switch proc.method {
		case QUERY:
			res, err = proc.queryHandler(c, queryParams)
			if err != nil {
				return err
			}
		case MUTATION:
			input, err := validateInput(c, proc)
			if err != nil {
				return err
			}
			res, err = proc.mutationHandler(c, queryParams, input)
			if err != nil {
				return err
			}
		}

		err = validateOutput(validatorFn, proc, res, fullRoute, MUTATION)
		if err != nil {
			return err
		}

		err = setHeaders(c, &res.Header)
		if err != nil {
			return err
		}
		return sendRes(c, res)
	}
	// if !proc.app.config.disableGenerateTS {
	// 	params := *new(queryParams)

	// 	input := *new(input)

	// 	output := *new(output)

	// 	genTypescript.AddProcedureToTree(fullRoute, params, input, output, genTypescript.Method(MUTATION))
	// }

	route.addProcedure(slug, &ProcedureInfo{
		method:      proc.method,
		handler:     fullHandler,
		validatorFn: route.getValidatorFn(),

		querySchema:  new(query),
		inputSchema:  new(input),
		outputSchema: new(output),
	})
	app := route.getApp()
	app.recalculateMux = true
}

func validateQuery[queryParams any, input any, output any](c *Ctx, proc *Procedure[queryParams, input, output], slug string) (queryParams, error) {

	queryParamInstance := new(queryParams)

	if proc.queryParamsSchema == nil || proc.validatorFn == nil {
		return *queryParamInstance, nil
	}

	if err := c.queryParser(queryParamInstance, slug); err != nil {
		return *queryParamInstance, err
	}
	if err := (*proc.validatorFn)(queryParamInstance); err != nil {

		return *queryParamInstance, &Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}

	}

	return *queryParamInstance, nil

}
func validateInput[queryParams any, input any, output any](c *Ctx, proc *Procedure[queryParams, input, output]) (input, error) {
	inputInstance := new(input)
	if proc.inputSchema == nil || proc.validatorFn == nil {
		return *inputInstance, nil
	}
	if err := c.BodyParser(inputInstance); err != nil {
		fmt.Println("err here at bodyParser of input", err.Error())
		return *inputInstance, err
	}
	// Validate the struct
	if err := (*proc.validatorFn)(inputInstance); err != nil {

		return *inputInstance, &Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}

	}
	return *inputInstance, nil
}
func validateOutput[queryParams any, input any, output any](validatorFn validatorFn, proc *Procedure[queryParams, input, output], res *Res[output], path string, method Method) error {
	if proc.outputSchema == nil || validatorFn == nil {
		return nil
	}

	if err := validatorFn(res.Body); err != nil {
		fmt.Printf(DefaultColors.Red+"INVALID OUTPUT ERROR at: %s, method : %s , error : %s \n", path, method, err.Error())
		return &Error{
			Code:    500,
			Message: "A server error has occurred. Please try again later",
		}
	}
	return nil
}
func setHeaders(ctx *Ctx, header *Header) error {
	if header.Authorization != "" {
		ctx.Set("Authorization", header.Authorization)
	}
	if header.CacheControl != "" {
		ctx.Set("Cache-Control", header.CacheControl)
	}
	if header.ContentEncoding != "" {
		ctx.Set("Content-Encoding", header.ContentEncoding)
	}
	if header.ContentType != "" {
		ctx.Set("Content-Type", header.ContentType)
	} else {
		header.ContentType = MIMEApplicationJSON
		ctx.Set("Content-Type", header.ContentType)
	}
	if header.Expires != "" {
		ctx.Set("Expires", header.Expires)
	}

	for _, cookie := range header.Cookies {
		if cookie != nil {
			ctx.Cookie(cookie)
		}
	}
	return nil
}
func sendRes[output any](ctx *Ctx, res *Res[output]) error {

	switch res.Header.ContentType {
	case MIMETextXML, MIMETextXMLCharsetUTF8:
		return ctx.XML(res.Body)
	case MIMETextPlain, MIMETextPlainCharsetUTF8:
		return ctx.SendString(fmt.Sprint(res.Body))

	case MIMEApplicationJSON, MIMEApplicationJSONCharsetUTF8:
		return ctx.JSON(res.Body)
	case MIMEApplicationJavaScript:
		return ctx.SendString(fmt.Sprint(res.Body))
	case MIMEApplicationForm:
		return ctx.SendString(fmt.Sprint(res.Body))
	case MIMEOctetStream:
		return ctx.SendString(fmt.Sprint(res.Body))
	case MIMEMultipartForm:
		ctx.SendString(fmt.Sprint(res.Body))
	default:
		return ctx.Status(400).SendString("Unsupported media type")
	}
	return nil
}
