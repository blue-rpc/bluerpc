package bluerpc

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Router struct {
	procedures  map[string]*ProcedureInfo
	app         *App
	routes      map[string]*Router
	mux         *http.ServeMux
	absPath     string
	mws         []Handler
	validatorFn *validatorFn
}

func (router *Router) getAbsPath() string {
	if router.absPath == "" {
		return "/"
	}
	return router.absPath

}

func (router *Router) addProcedure(slug string, info *ProcedureInfo) {
	router.procedures[slug] = info
}
func (router *Router) getValidatorFn() *validatorFn {
	return router.validatorFn
}
func (router *Router) getApp() *App {
	return router.app
}

// changes the validator function for all of the connected procedures unless those procedures directly have validator functions set
func (r *Router) Validator(fn validatorFn) {
	r.validatorFn = &fn
}

// Creates a new route or returns an existing one if that route was already created
func (r *Router) Router(slug string) *Router {
	// Cannot have an empty prefix
	if slug == "" {
		panic("no relative path provided")
	}
	// Prefix always start with a '/' or '*'
	if slug[0] != '/' {
		slug = "/" + slug
	}

	// Check if the key exists in the map
	route, exists := r.routes[slug]
	if exists {
		// Return the existing route
		return route
	}

	//this handles the case where the user passes in a route with multiple slashes. Something like /users/image/info
	//it just creates all of the necessary routes up to /info in the background
	slugs, err := splitStringOnSlash(slug)
	if err != nil {
		panic(err)
	}
	currentRoute := r
	if len(slugs) > 1 {
		lastSlugIndex := len(slugs) - 1
		for i := 0; i < lastSlugIndex; i++ {
			prevRoute := currentRoute
			currentRoute = prevRoute.Router(slugs[i])
		}

	}

	newRouter := &Router{
		absPath:     r.absPath + slug,
		mux:         http.NewServeMux(),
		routes:      map[string]*Router{},
		procedures:  map[string]*ProcedureInfo{},
		mws:         []Handler{},
		validatorFn: r.validatorFn,
		app:         r.app,
	}
	if strings.ContainsRune(slug, ':') {
		panic("You are not allowed to create dynamic routes. Read the docs from here :")
	}
	currentRoute.routes[slug] = newRouter

	return newRouter
}

// prefix is the ROUTE PREFIX
// root is the ROOT folder
func (r *Router) Static(prefix, root string, config ...*Static) {

	if root == "" {
		root = "."
	}
	// Cannot have an empty prefix
	if prefix == "" {
		prefix = "/"
	}

	// Prefix always start with a '/' or '*'
	if prefix[0] != '/' {
		prefix = "/" + prefix
	}

	var actualConfig *Static

	if len(config) > 0 {
		actualConfig = config[0]
	} else {
		actualConfig = &Static{}
	}
	if actualConfig.CacheDuration == 0 {
		actualConfig.CacheDuration = time.Second * 10
	}

	if actualConfig.Index == "" {
		actualConfig.Index = "/index.html"
	}
	if actualConfig.Index[0] != '/' {
		actualConfig.Index = "/" + actualConfig.Index
	}
	slugs, _ := splitStringOnSlash(prefix)

	//if the prefix is not "/" then loop through each sub routes until you get to "/", then attach the static method to the last "/"
	if prefix != "/" {
		loopRoute := r
		for i := 0; i < len(slugs); i++ {
			prevRoute := loopRoute
			loopRoute = prevRoute.Router(slugs[i])
		}

		loopRoute.Static("/", root)
		return
	}
	fmt.Println("absolute route where create static function is called", r.getAbsPath())
	r.addProcedure("/", &ProcedureInfo{
		method:      QUERY,
		validatorFn: r.validatorFn,
		handler:     createStaticFunction(prefix, root, actualConfig),
	})
}

func (r *Router) Use(middlewares ...Handler) {
	if len(middlewares) == 0 {
		panic("Use called without any middleware arguments")
	}

	r.mws = append(r.mws, middlewares...)

}

// prints all of the info for this route, including its procedures attached and all of its nested routes
func (r *Router) PrintInfo() {

	absPath := r.getAbsPath()
	if absPath == "" {
		absPath = "/"
	}
	generalInfo := fmt.Sprintf("%s Route : %s %s", DefaultColors.Green, absPath, DefaultColors.Reset)

	if len(r.procedures) == 0 {
		generalInfo += fmt.Sprintf("%s No Procedures %s", DefaultColors.Red, DefaultColors.Reset)
	}
	fmt.Println(generalInfo)
	for slug, procInfo := range r.procedures {

		pathAndMethod := fmt.Sprintf("	%s : %s ", procInfo.method, slug)
		inputsAndOutputs := &strings.Builder{}

		inputsAndOutputs.WriteString("(")

		switch procInfo.method {
		case QUERY:

			queryType := getType(procInfo.querySchema)

			inputsAndOutputs.WriteString(goToTsObj(queryType))
		case MUTATION:
			inputsAndOutputs.WriteString("{")
			inputsAndOutputs.WriteString("query:")
			queryType := getType(procInfo.querySchema)
			inputsAndOutputs.WriteString(goToTsObj(queryType))
			inputsAndOutputs.WriteString(",")
			inputType := getType(procInfo.inputSchema)
			inputsAndOutputs.WriteString(goToTsObj(inputType))
			inputsAndOutputs.WriteString("}")

		}

		inputsAndOutputs.WriteString(")=>")

		outputType := getType(procInfo.outputSchema)
		inputsAndOutputs.WriteString(goToTsObj(outputType))

		fmt.Println(pathAndMethod + inputsAndOutputs.String())
	}

	for _, nestedRoute := range r.routes {
		nestedRoute.PrintInfo()
	}
}

func createCtx(w http.ResponseWriter, r *http.Request) *Ctx {
	return &Ctx{
		httpW: w,
		httpR: r,
	}

}
func methodsMatch(httpMethod string, bluerpcMethod Method) bool {
	switch bluerpcMethod {
	case QUERY:
		return httpMethod == "GET"
	case MUTATION:
		return httpMethod == "POST"
	}
	return false
}
