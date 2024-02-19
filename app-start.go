package bluerpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/pprof"
	"os"
	"strings"
	"time"
)

func (a *App) Listen(port string) error {
	a.port = port

	if a.recalculateMux {

		nestedMux, totalRoutes := buildMux(a.startRoute, a.startRoute.mws, 0)
		if !a.config.DisableGenerateTS {
			generateTs(a)
		}
		a.serveMux = &http.ServeMux{}
		a.serveMux.Handle("/", nestedMux)

		if a.config.EnablePProf {
			attachPprofRoutes(a.serveMux)
		}
		if !a.config.DisableInfoPrinting {
			var serverUrl string
			if a.config.ServerURL == "" {
				serverUrl = "http://127.0.0.1"
			} else {
				serverUrl = a.config.ServerURL
			}

			serverUrl += port
			printStartServerInfo(totalRoutes, serverUrl)
			a.PrintRoutes()
		}

		a.recalculateMux = false
	}

	a.server = &http.Server{
		Addr:    port,
		Handler: a.serveMux,
	}

	return func() error {
		var err error
		if a.config.SSLCertPath != "" {
			err = a.server.ListenAndServeTLS(a.config.SSLCertPath, a.config.SSLKey)
		} else {
			err = a.server.ListenAndServe()
		}

		if err != nil {
			if err == http.ErrServerClosed {
				// Server closed gracefully, not an error
				return nil
			}
			// Actual error
			return err
		}
		return nil
	}()

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

	for slug, route := range router.routes {
		localRoute := route
		localSlug := slug
		nextMws := append(prevMws, route.mws...)
		nestedMux, newTotalRoutes := buildMux(localRoute, nextMws, totalRoutes)
		totalRoutes = newTotalRoutes
		mux.Handle(slug+"/", http.StripPrefix(localSlug, nestedMux))
	}
	setupProcedures(router.getAbsPath(), mux, router.procedures, router.mws)

	return mux, totalRoutes
}

// AttachPprofRoutes adds the pprof routes to the provided mux.
func attachPprofRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	// Register other pprof endpoints by specifying the profile name as the parameter to pprof.Handler.
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	// You can add more handlers here based on the pprof documentation.
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
	ctx, cancel := context.WithTimeout(context.Background(), 5)
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
