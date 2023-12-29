package bluerpc

type Error struct {
	Code    int
	Message string
}

func (bluerpcErr *Error) Error() string {
	return bluerpcErr.Message
}
