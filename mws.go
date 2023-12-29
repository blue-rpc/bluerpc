package bluerpc

// Simple middleware that always returns json
func DefaultErrorMiddleware(ctx *Ctx) error {

	err := ctx.Next()

	if err == nil {
		return nil
	}

	if e, ok := err.(*Error); ok {

		return ctx.Status(e.Code).JSON(Map{
			"message": e.Message,
		})
	}
	return ctx.Status(500).JSON(Map{
		"message": err.Error(),
	})
}
