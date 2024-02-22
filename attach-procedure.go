package bluerpc

import (
	"fmt"
	"net/http"
	"strings"
)

type Route interface {
	getAbsPath() string
	getAuthorizer() *Authorizer
	isAuthorized() bool
	addProcedure(string, *ProcedureInfo)
	getValidatorFn() *validatorFn
	getApp() *App
	Router(string) *Router
}

func (proc *Procedure[query, input, output]) Attach(route Route, slug string) {

	if route.isAuthorized() {
		proc = proc.Protected()
	}
	if proc.authorizer == nil {
		proc.authorizer = route.getAuthorizer()
	}

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

	if absPath == "/" {
		absPath = ""
	}
	dynamicSlugs := findDynamicSlugs(slug)
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
			err := checkContentTypeValidity(c.httpR.Header.Get("Content-Type"), proc.acceptedContentType)
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

		err = validateOutput(proc, res, fullRoute, MUTATION)
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
		method:       proc.method,
		handler:      fullHandler,
		validatorFn:  route.getValidatorFn(),
		dynamicSlugs: dynamicSlugs,
		querySchema:  new(query),
		inputSchema:  new(input),
		outputSchema: new(output),
		protected:    proc.protected,
		authorizer:   proc.authorizer,
	})
	app := route.getApp()

	app.recalculateMux = true
}

func validateQuery[query any, input any, output any](c *Ctx, proc *Procedure[query, input, output], slug string) (query, error) {
	queryParamInstance := new(query)

	if !proc.hasQuery {
		return *queryParamInstance, nil
	}

	if err := c.queryParser(queryParamInstance, slug); err != nil {
		return *queryParamInstance, err
	}

	validatorFn := *proc.validatorFn
	if validatorFn == nil {
		return *queryParamInstance, nil
	}
	if err := validatorFn(queryParamInstance); err != nil {
		return *queryParamInstance, &Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}

	}

	return *queryParamInstance, nil

}
func validateInput[query any, input any, output any](c *Ctx, proc *Procedure[query, input, output]) (input, error) {
	inputInstance := new(input)
	if !proc.hasInput {
		return *inputInstance, nil
	}

	if err := c.bodyParser(inputInstance); err != nil {
		return *inputInstance, err
	}

	validatorFn := *proc.validatorFn
	if validatorFn == nil {
		return *inputInstance, nil
	}
	// Validate the struct
	if err := validatorFn(inputInstance); err != nil {

		return *inputInstance, &Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}

	}
	return *inputInstance, nil
}
func validateOutput[query any, input any, output any](proc *Procedure[query, input, output], res *Res[output], path string, method Method) error {

	validatorFn := *proc.validatorFn
	if !proc.hasOutput || validatorFn == nil {
		return nil
	}
	if err := validatorFn(res.Body); err != nil {
		fmt.Println("error", err.Error())
		fmt.Printf(DefaultColors.Red+"INVALID OUTPUT ERROR at: %s, method : %s , error : %s \n", path, method, err.Error()+DefaultColors.Reset)
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
			ctx.SetCookie(cookie)
		}
	}
	return nil
}
func sendRes[output any](ctx *Ctx, res *Res[output]) error {

	status := res.Header.Status
	if res.Header.Status == 0 {
		status = 200
	}
	switch res.Header.ContentType {
	case TextXML, TextXMLCharsetUTF8:
		return ctx.status(status).xML(res.Body)
	case TextPlain, TextPlainCharsetUTF8:
		return ctx.status(status).SendString(fmt.Sprint(res.Body))
	case ApplicationJSON, ApplicationJSONCharsetUTF8:
		return ctx.status(status).jSON(res.Body)
	case ApplicationJavaScript:
		return ctx.status(status).SendString(fmt.Sprint(res.Body))
	case ApplicationForm:
		return ctx.status(status).SendString(fmt.Sprint(res.Body))
	case OctetStream:
		return ctx.status(status).SendString(fmt.Sprint(res.Body))
	case MultipartForm:
		ctx.SendString(fmt.Sprint(res.Body))
	default:
		return ctx.status(400).SendString("Unsupported media type")
	}
	return nil
}

func checkContentTypeValidity(contentType string, validContentTypes []string) error {

	if contentType == "" || len(validContentTypes) == 0 {
		return nil
	}
	fullContentTypes := ""

	for i, validType := range validContentTypes {
		if strings.Contains(contentType, validType) {
			return nil
		}
		fullContentTypes += validType
		if i != len(validContentTypes)-1 {
			fullContentTypes += ", "
		}

	}

	return &Error{
		Code:    400,
		Message: fmt.Sprintf("invalid content type. server only accepts %s", fullContentTypes),
	}

}
