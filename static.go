package bluerpc

import (
	"fmt"
	"net/http"
	"os"
	"path"
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
	return func(ctx *Ctx) error {
		// Split the path into segments using "/"
		segments := strings.Split(ctx.httpR.URL.Path, "/")
		slug := "/"

		if len(segments) >= 2 {
			slug += segments[len(segments)-1]
		}

		if config.Download {
			ctx.Set("Content-Disposition", "attachment")
		}
		if config.Next != nil && config.Next(ctx) {
			return nil
		}
		if config.MaxAge > 0 {
			ctx.httpW.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", config.MaxAge))
		}
		if slug == prefix || segments[len(segments)-1] == "" {
			if _, err := os.Stat(root + config.Index); err != nil {
				return err
			}

			http.ServeFile(ctx.httpW, ctx.httpR, root+config.Index)
		} else {
			filePath := path.Join(root, path.Base(ctx.httpR.URL.Path))
			http.ServeFile(ctx.httpW, ctx.httpR, filePath)
		}
		return nil
	}
}
