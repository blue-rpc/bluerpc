package bluerpc

import (
	"net/http"
)

type Handler = func(ctx *Ctx) error
type AuthHandler func(ctx *Ctx) (any, error)

type Res[T any] struct {
	Header Header
	Body   T
}
type Header struct {
	Status          int
	Authorization   string         // Credentials for authenticating the client to the server
	CacheControl    string         // Directives for caching mechanisms in both requests and responses
	ContentEncoding string         // The encoding of the body
	ContentType     string         // The MIME type of the body of the request (used with POST and PUT requests)
	Expires         string         // Gives the date/time after which the response is considered stale
	Cookies         []*http.Cookie //Cookies
}

// First Generic argument is QUERY PARAMS.
// Second is OUTPUT
type Query[query any, output any] func(ctx *Ctx, query query) (*Res[output], error)

// First Generic argument is QUERY PARAMETERS.
// Second is INPUT.
// Third is OUTPUT.
type Mutation[query any, input any, output any] func(ctx *Ctx, query query, input input) (*Res[output], error)

type ErrorResponse struct {
	Message string `json:"message"`
}

func (c *Ctx) Next() error {
	err := c.nextHandler(c)
	c.nextHandler = nil

	if err != nil {
		return err
	}

	return nil

}
