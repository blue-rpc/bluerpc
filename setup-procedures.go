package bluerpc

import (
	"fmt"
	"net/http"
	"strings"
)

func setupProcedures(absPath string, mux *http.ServeMux, procs map[string]*ProcedureInfo, mws []Handler) {

	// Temporary variable to hold whether the root handler is provided
	var rootProc *ProcedureInfo
	rootProvided := false

	// Check if the root handler is provided in the map
	if proc, ok := procs["/"]; ok {
		rootProc = proc
		rootProvided = true
	}

	// First, set up all routes except for those matching "/:" and "/"
	for path, proc := range procs {
		localProc := proc
		localPath := path
		if localPath == "/" {
			continue
		}

		if !strings.HasPrefix(localPath, "/:") && localPath != "/" {
			attachProcedureToMux(mux, localPath, localProc, mws)
		} else {
			rootProc = localProc
			if rootProvided {
				panic(fmt.Sprintf("You have at least two procedures at this localPath %s that are dynamic (either `/` or that start with `/:`). You can only have 1", absPath))
			}
			rootProvided = true

		}
	}

	if rootProvided {
		attachProcedureToMux(mux, "/", rootProc, mws)
	}

}

func attachProcedureToMux(mux *http.ServeMux, slug string, proc *ProcedureInfo, mws []Handler) {
	mux.HandleFunc(slug, func(w http.ResponseWriter, r *http.Request) {
		ctx := createCtx(w, r)
		var allHandlersArray []Handler
		allHandlersArray = append(allHandlersArray, mws...)
		if !methodsMatch(r.Method, proc.method) {
			allHandlersArray = append(allHandlersArray, func(Ctx *Ctx) error {
				return &Error{
					Code:    405,
					Message: "Method not allowed",
				}
			})
			fullHandler := generateFullHandler(allHandlersArray)
			fullHandler(ctx)
			return
		}

		if proc.protected {
			if proc.authorizer == nil || proc.authorizer.Handler == nil {
				panic("You have set a procedure to be protected and yet you did not provide any authorization functions to either its router parents or the app.")
			}
			allHandlersArray = append(allHandlersArray, func(Ctx *Ctx) error {
				authRes, err := proc.authorizer.Handler(ctx)

				if err != nil {
					return &Error{
						Code:    401,
						Message: err.Error(),
					}
				}
				ctx.auth = authRes
				fmt.Println("ctx auth", ctx.auth)
				return nil
			})
		}

		allHandlersArray = append(allHandlersArray, proc.handler)

		fullHandler := generateFullHandler(allHandlersArray)
		fullHandler(ctx)

	})
}

// Create a chain function where you run each middleware in a recursive matter
func generateFullHandler(handlers []Handler) Handler {

	if len(handlers) == 0 {
		return func(ctx *Ctx) error {
			return ctx.status(404).jSON(Map{"message": "not found"})
		}
	}
	chain := handlers[len(handlers)-1]

	//this loops from the end of the array to the start.
	for i := len(handlers) - 2; i >= 0; i-- {
		//Start at the end of the array. For each step
		currentIndex := i
		outsideChain := chain
		chain = func(ctx *Ctx) error {
			// set the next function to be the previous chain functions combined
			ctx.nextHandler = outsideChain
			//run the given handler

			err := handlers[currentIndex](ctx)

			if err != nil {
				return err
			}
			// if Next() was not run by the user then run Next() to continue
			if ctx.nextHandler != nil {
				return ctx.Next()
			}
			return nil

		}
	}
	return chain
}
