package bluerpc

import (
	"net/http"
	"time"
)

type Router struct {
	procedures  map[string]*ProcedureInfo
	app         *App
	Routes      map[string]*Router
	mux         *http.ServeMux
	absPath     string
	mws         []Handler
	validatorFn *validatorFn
}

func (router *Router) getAbsPath() string {
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
func (r *Router) Router(relativePath string, procedures ...map[string]Procedure[any, any, any]) *Router {
	// Cannot have an empty prefix
	if relativePath == "" {
		panic("no relative path provided")
	}
	// Prefix always start with a '/' or '*'
	if relativePath[0] != '/' {
		relativePath = "/" + relativePath
	}
	newRouter := &Router{
		absPath:     r.absPath + relativePath,
		mux:         http.NewServeMux(),
		Routes:      map[string]*Router{},
		procedures:  map[string]*ProcedureInfo{},
		mws:         []Handler{},
		validatorFn: r.validatorFn,
		app:         r.app,
	}
	r.Routes[relativePath] = newRouter
	var proceduresMap map[string]Procedure[any, any, any]
	if len(procedures) > 0 {
		proceduresMap = procedures[0]
	}
	for route, procedure := range proceduresMap {
		procedure.Attach(newRouter, route)
	}
	return newRouter
}
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
	// Strip trailing slashes from the root path
	if len(root) > 0 && root[len(root)-1] == '/' {
		root = root[:len(root)-1]
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
	staticRouter := r.Router(prefix)
	r.procedures[prefix] = &ProcedureInfo{
		method:      QUERY,
		validatorFn: r.validatorFn,
		handler: func(ctx *Ctx) error {
			http.Redirect(ctx.httpW, ctx.httpR, prefix+"/", http.StatusMovedPermanently)
			return nil
		},
	}
	staticRouter.addProcedure("/", &ProcedureInfo{
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

func createCtx(w http.ResponseWriter, r *http.Request) *Ctx {
	return &Ctx{
		httpW:        w,
		httpR:        r,
		indexHandler: 0,
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
