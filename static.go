package bluerpc

import (
	"net/http"
	"strings"
	"time"
)

type Static struct {
	// When set to true, enables direct download.
	// Optional. Default value false.
	Download bool `json:"download"`

	// The name of the index file for serving a directory.
	// Optional. Default value "index.html".
	Index string `json:"index"`

	// Expiration duration for inactive file handlers.
	// Use a negative time.Duration to disable it.
	//
	// Optional. Default value 10 * time.Second.
	CacheDuration time.Duration `json:"cache_duration"`

	// The value for the Cache-Control HTTP-header
	// that is set on the file response. MaxAge is defined in seconds.
	//
	// Optional. Default value 0.
	MaxAge int `json:"max_age"`

	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(c *Ctx) bool
}

func createStaticFunction(prefix, root string, config *Static) func(c *Ctx) error {
	fs := http.FileServer(http.Dir(root))

	return func(ctx *Ctx) error {

		if strings.Contains(ctx.httpR.URL.Path, ".") {
			// Let the file server handle the request
			fs.ServeHTTP(ctx.httpW, ctx.httpR)
		} else {
			// For any other requests, serve index.html
			http.ServeFile(ctx.httpW, ctx.httpR, root+config.Index)
		}

		return nil
	}

}
