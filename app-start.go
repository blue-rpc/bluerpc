package bluerpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

func (a *App) Listen(port string) error {
	if a.recalculateMux {

		nestedMux, totalRoutes := buildMux(a.startRoute, a.startRoute.mws, 0)
		if !a.config.DisableGenerateTS {
			generateTs(a)
		}
		a.serveMux = &http.ServeMux{}
		a.serveMux.Handle("/", nestedMux)

		if !a.config.DisableInfoPrinting {
			var serverUrl string
			if a.config.ServerURL == "" {
				serverUrl = "http://127.0.0.1"
			} else {
				serverUrl = a.config.ServerURL
			}

			serverUrl += port
			printStartServerInfo(totalRoutes, serverUrl)
		}

		a.recalculateMux = false
	}

	a.port = port

	a.server = &http.Server{
		Addr:    port,
		Handler: a.serveMux,
	}
	func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server Crashed : %v", err)
		}
	}()

	return nil
}
func buildMux(router *Router, prevMws []Handler, totalRoutes int) (*http.ServeMux, int) {

	mux := &http.ServeMux{}
	router.mux = mux

	// mux.HandleFunc("/hello-world", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("hello world"))
	// })
	if router.mws == nil {
		router.mws = prevMws
	} else {
		router.mws = append(router.mws, prevMws...)
	}
	totalRoutes += len(router.procedures)
	for slug, proc := range router.procedures {

		localSlug := slug
		localProc := proc

		// if the procedure slug starts with : that means this is a dynamic route. we store the slug inside of the dynamic slug in order for us to store that variable for later use when validating the query params
		if strings.HasPrefix(localSlug, "/:") {
			localSlug = "/"
		}
		mux.HandleFunc(localSlug, func(w http.ResponseWriter, r *http.Request) {
			ctx := createCtx(w, r)
			var allHandlersArray []Handler
			if methodsMatch(r.Method, localProc.method) {

				allHandlersArray = append(router.mws, localProc.handler)
			} else {
				allHandlersArray = append(router.mws, func(Ctx *Ctx) error {
					return fmt.Errorf("Method not allowed")
				})
			}
			fullHandler := generateFullHandler(allHandlersArray)
			fullHandler(ctx)

		})
	}

	for slug, route := range router.Routes {
		nextMws := append(prevMws, route.mws...)
		nestedMux, newTotalRoutes := buildMux(route, nextMws, totalRoutes)
		totalRoutes = newTotalRoutes
		mux.Handle(slug+"/", http.StripPrefix(slug, nestedMux))
	}
	return mux, totalRoutes
}

// Create a chain function where you run each middleware in a recursive matter
func generateFullHandler(handlers []Handler) Handler {

	if len(handlers) == 0 {
		return func(ctx *Ctx) error {
			return ctx.Status(404).JSON(Map{"message": "not found"})
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

			handlers[currentIndex](ctx)
			// if Next() was not run by the user then run Next() to continue
			if ctx.nextHandler != nil {
				return ctx.Next()
			}
			return nil

		}
	}
	return chain
}
func (a *App) Test(req *http.Request, port ...string) (*http.Response, error) {
	userPort := ":8080"
	if len(port) > 1 {
		userPort = port[0]
	}
	if a.port == "" {
		go a.Listen(userPort)
	}

	if err := waitForServerReady(userPort); err != nil {
		return nil, err
	}
	time.Sleep(1 * time.Second)

	// Create a new response recorder to capture the response.
	rr := httptest.NewRecorder()

	// Serve the request using the app's serveMux.
	a.serveMux.ServeHTTP(rr, req)

	a.Shutdown()
	// Return the recorded response.
	return rr.Result(), nil
}
func (a *App) Shutdown() {

	a.mutex.Lock()
	defer a.mutex.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	if err := a.server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	a.port = ""

}
func waitForServerReady(port string) error {
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for server to start on port %s", port)
		case <-ticker.C:
			conn, err := net.Dial("tcp", port)
			if err == nil {
				conn.Close()
				return nil
			}
		}
	}
}

func printStartServerInfo(numProcesses int, address string) {
	const colorStart = "\033[38;2;52;211;153m" // ANSI escape code for #34d399
	pid := os.Getpid()

	const borderLength = 60
	border := colorStart + strings.Repeat("_", borderLength) + DefaultColors.Reset
	emptyLine := colorStart + "|" + strings.Repeat(" ", borderLength-2) + "|" + DefaultColors.Reset
	fmt.Println(border)
	// Each line of the ASCII art is individually centered
	industrialBlueRPC := []string{
		"  ____  _              ____  ____   _____ ",
		" | __ )| |_   _  ___  |  _ \\|  _ \\ / ____|",
		" |  _ \\| | | | |/ _ \\ | |_) | |_) | |     ",
		" | |_) | | |_| |  __/ |  __/|  _ <| |____ ",
		" |____/|_|\\__,_|\\___| |_|   |_| \\_\\\\_____|",
	}

	// Center each line of the ASCII art and print it
	for _, line := range industrialBlueRPC {
		fmt.Println(colorStart + "|" + centerText(line, borderLength-2) + "|" + DefaultColors.Reset)
	}
	fmt.Println(emptyLine)

	fmt.Println(emptyLine)
	fmt.Printf(colorStart+"| %s%s|\n"+DefaultColors.Reset, leftAlignText(fmt.Sprintf("Server Started on %s", address), borderLength-4), " ")
	fmt.Printf(colorStart+"| %s%s|\n"+DefaultColors.Reset, leftAlignText(fmt.Sprintf("Number of Processes: %d", numProcesses), borderLength-4), " ")
	fmt.Printf(colorStart+"| %s%s|\n"+DefaultColors.Reset, leftAlignText(fmt.Sprintf("PID: %d", pid), borderLength-4), " ")
	fmt.Println(border)

}

func centerText(text string, width int) string {
	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", padding)
}

func leftAlignText(text string, width int) string {
	return text + strings.Repeat(" ", width-len(text))
}
