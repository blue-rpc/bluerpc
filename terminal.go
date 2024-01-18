package bluerpc

// DefaultColors contains ANSI escape codes for default colors.
var DefaultColors = struct {
	Red   string
	Green string
	Reset string
}{
	Red:   "\x1b[31m",
	Green: "\x1b[32m",
	Reset: "\x1b[0m",
}
