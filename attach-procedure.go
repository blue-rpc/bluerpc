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
	Router(string) *Router
}

func (proc *Procedure[query, input, output]) Attach(route Route, slug string) {

	//if the user attached a nested route, something like /users/images/info, this part of the function will split the slug into its corresponding parts, will create the needed nested routes and then will attach the procedure to the last created route
	// in the case of /users/images/info it will create a /users route, in /users it will create /images and it will attach the procedure on images at /info
	slugs, err := splitStringOnSlash(slug)

	if err != nil {
		panic(err)
	}

	if len(slugs) > 1 {
		lastSlugIndex := len(slugs) - 1

		loopRoute := route
		for i := 0; i < lastSlugIndex; i++ {
			prevRoute := loopRoute
			loopRoute = prevRoute.Router(slugs[i])
		}
		proc.Attach(loopRoute, slugs[lastSlugIndex])
		return
	}

	absPath := route.getAbsPath()
	validatorFn := *proc.validatorFn
	fullRoute := absPath + slug
	fullHandler := func(c *Ctx) error {

		query, err := validateQuery(c, proc, slug)
		if err != nil {
			return err
		}
		var res *Res[output]

		switch proc.method {
		case QUERY:
			res, err = proc.queryHandler(c, query)
			if err != nil {
				return err
			}
		case MUTATION:
			err := checkContentTypeValidity(c.httpW.Header().Get("Content-Type"), proc.acceptedContentType)
			if err != nil {
				return err
			}
			input, err := validateInput(c, proc)
			if err != nil {
				return err
			}
			res, err = proc.mutationHandler(c, query, input)
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

	route.addProcedure(slug, &ProcedureInfo{
		method:      proc.method,
		handler:     fullHandler,
		validatorFn: route.getValidatorFn(),

		querySchema:  proc.querySchema,
		inputSchema:  proc.inputSchema,
		outputSchema: proc.outputSchema,
	})
	app := route.getApp()

	app.recalculateMux = true
}

func validateQuery[query any, input any, output any](c *Ctx, proc *Procedure[query, input, output], slug string) (query, error) {

	queryParamInstance := new(query)

	if proc.querySchema == nil || proc.validatorFn == nil {
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
func validateInput[query any, input any, output any](c *Ctx, proc *Procedure[query, input, output]) (input, error) {
	inputInstance := new(input)
	if proc.inputSchema == nil || proc.validatorFn == nil {
		return *inputInstance, nil
	}
	if err := c.bodyParser(inputInstance); err != nil {
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
func validateOutput[query any, input any, output any](validatorFn validatorFn, proc *Procedure[query, input, output], res *Res[output], path string, method Method) error {
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
		header.ContentType = ApplicationJSON
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
	case TextXML, TextXMLCharsetUTF8:
		return ctx.xML(res.Body)
	case TextPlain, TextPlainCharsetUTF8:
		return ctx.SendString(fmt.Sprint(res.Body))

	case ApplicationJSON, ApplicationJSONCharsetUTF8:
		return ctx.jSON(res.Body)
	case ApplicationJavaScript:
		return ctx.SendString(fmt.Sprint(res.Body))
	case ApplicationForm:
		return ctx.SendString(fmt.Sprint(res.Body))
	case OctetStream:
		return ctx.SendString(fmt.Sprint(res.Body))
	case MultipartForm:
		ctx.SendString(fmt.Sprint(res.Body))
	default:
		return ctx.status(400).SendString("Unsupported media type")
	}
	return nil
}

func checkContentTypeValidity(contentType string, validContentTypes []string) error {

	fullContentTypes := ""
	for i, validType := range validContentTypes {
		if validType == contentType {
			return nil
		}
		fullContentTypes += validType
		if i != len(validContentTypes)-1 {
			fullContentTypes += ", "
		}

	}

	return fmt.Errorf("invalid content type. server only accepts %s", fullContentTypes)
}
