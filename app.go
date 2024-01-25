package bluerpc

import (
	"fmt"
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

	newApp := App{
		config:         cfg,
		serveMux:       http.NewServeMux(),
		mutex:          sync.Mutex{},
		recalculateMux: true,
	}

	mws := []Handler{}
	if cfg.CORS_Origin != "" {
		corsMw := createDefaultCorsOrigin(cfg.CORS_Origin)
		mws = append(mws, corsMw)
	}
	mws = append(mws, cfg.ErrorMiddleware)

	startRouter := &Router{
		routes:      map[string]*Router{},
		procedures:  map[string]*ProcedureInfo{},
		mux:         http.NewServeMux(),
		validatorFn: &cfg.ValidatorFn,
		absPath:     "",
		mws:         mws,
		app:         &newApp,
		authorizer:  cfg.Authorizer,
	}

	newApp.startRoute = startRouter
	return &newApp
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
	// Prefix must always start with a '/'
	if relativePath[0] != '/' {
		relativePath = "/" + relativePath
	}

	return a.startRoute.Router(relativePath)

}
func (a *App) getAuthorizer() *Authorizer {
	return a.config.Authorizer
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
func (a *App) PrintRoutes() {
	fmt.Println("")
	fmt.Println("_____________________________________________")
	a.startRoute.PrintInfo()
	fmt.Println("_____________________________________________")
	fmt.Println("")

}
