package bluerpc

// Simple middleware that always returns json
func DefaultErrorMiddleware(ctx *Ctx) error {
	err := ctx.Next()

	if err == nil {
		return nil
	}

	if e, ok := err.(*Error); ok {

		return ctx.status(e.Code).jSON(Map{
			"message": e.Message,
		})
	}
	return ctx.status(500).jSON(Map{
		"message": err.Error(),
	})
}

func createDefaultCorsOrigin(cors string) func(ctx *Ctx) error {
	return func(ctx *Ctx) error {
		ctx.httpW.Header().Set("Access-Control-Allow-Origin", cors)
		ctx.httpW.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		return nil
	}
}
