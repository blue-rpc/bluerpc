package bluerpc

import (
	"net/http"
	"sync"
)

type validatorFn func(interface{}) error

type App struct {
	startRoute     *Router
	config         *Config
	port           string
	serveMux       *http.ServeMux
	server         *http.Server
	mutex          sync.Mutex
	recalculateMux bool
}

func New(blueConfig ...*Config) *App {

	cfg := setAppDefaults(blueConfig)

	startRouter := &Router{
		routes:      map[string]*Router{},
		procedures:  map[string]*ProcedureInfo{},
		mux:         http.NewServeMux(),
		validatorFn: &cfg.ValidatorFn,
		absPath:     "/",
		mws:         []Handler{cfg.ErrorMiddleware},
	}

	return &App{
		config:         cfg,
		serveMux:       http.NewServeMux(),
		startRoute:     startRouter,
		mutex:          sync.Mutex{},
		recalculateMux: true,
	}
}

func setAppDefaults(blueConfig []*Config) *Config {

	var cfg *Config

	if len(blueConfig) > 0 {
		cfg = blueConfig[0]
	} else {
		cfg = &Config{}
	}

	var (
		tsOutputPath = "./output.ts"
	)

	if cfg.OutputPath == "" {
		cfg.OutputPath = tsOutputPath
	}
	if cfg.ErrorMiddleware == nil {
		cfg.ErrorMiddleware = DefaultErrorMiddleware

	}

	return cfg
}
func (a *App) Router(relativePath string) *Router {
	// Cannot have an empty prefix
	if relativePath == "" {
		panic("no relative path provided")
	}
	// Prefix always start with a '/' or '*'
	if relativePath[0] != '/' {
		relativePath = "/" + relativePath
	}
	newRouter := &Router{
		absPath:     relativePath,
		mux:         http.NewServeMux(),
		routes:      map[string]*Router{},
		procedures:  map[string]*ProcedureInfo{},
		mws:         []Handler{},
		validatorFn: &a.config.ValidatorFn,
		app:         a,
	}
	a.startRoute.routes[relativePath] = newRouter
	return newRouter

}

func (a *App) getAbsPath() string {
	return "/"
}
func (a *App) addProcedure(slug string, info *ProcedureInfo) {
	a.startRoute.procedures[slug] = info

}
func (a *App) getValidatorFn() *validatorFn {
	return &a.config.ValidatorFn
}
func (a *App) getApp() *App {
	return a
}

func (a *App) Use(middleware Handler) *App {
	a.startRoute.Use(middleware)
	return a
}
func (a *App) Static(prefix, root string, config ...*Static) {
	a.startRoute.Static(prefix, root, config...)
}
