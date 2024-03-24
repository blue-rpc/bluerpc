package bluerpc

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestStatic(t *testing.T) {
	fmt.Println(DefaultColors.Green + "TESTING STATIC METHOD" + DefaultColors.Reset)

	validate := validator.New(validator.WithRequiredStructEnabled())
	app := New(&Config{
		ValidatorFn:         validate.Struct,
		DisableInfoPrinting: true,
		DisableGenerateTS:   true,
	})
	dir, err := os.MkdirTemp("", "local_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(dir) // Clean up.
	app.Static("/test", "local_test")

	const htmlBody = "<html><body>Hello World</body></html>"
	const cssBody = "body { color: red; }"
	createFile(t, dir, "index.html", "<html><body>Hello World</body></html>")
	createFile(t, dir, "assets.css", "body { color: red; }")

	// Start a test HTTP server.
	server := httptest.NewServer(http.FileServer(http.Dir(dir)))
	defer server.Close()
	// Test requests to index.html and assets.css.

	htmlReq, err := http.NewRequest("GET", "http://localhost:8080/test/index.html", nil)
	if err != nil {
		t.Fatal(err)
	}
	cssReq, err := http.NewRequest("GET", "http://localhost:8080/test/assets.css", nil)
	if err != nil {
		t.Fatal(err)
	}

	htmlRes, err := app.Test(htmlReq, ":8080")
	if err != nil {
		t.Fatal(err)
	}
	cssRes, err := app.Test(cssReq, ":8080")
	if err != nil {
		t.Fatal(err)
	}
	testBody := func(res *http.Response, expectedBody string) bool {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)

		}
		return string(body) == expectedBody
	}
	testBody(htmlRes, htmlBody)
	testBody(cssRes, cssBody)

	fmt.Println(DefaultColors.Green + "PASSED TESTING STATIC METHOD" + DefaultColors.Reset)

}
func TestStaticOutput(t *testing.T) {
	fmt.Println(DefaultColors.Green + "TESTING STATIC OUTPUT" + DefaultColors.Reset)
	app := New()
	app.Static("/assets", "")
	builder := strings.Builder{}
	nodeToTS(&builder, app.startRoute, true, "")
	if builder.String() != "{[`assets`]:{}}" {
		t.Fail()
	}
	fmt.Println(DefaultColors.Green + "PASSED TESTING STATIC OUTPUT" + DefaultColors.Reset)
}
func createFile(t *testing.T, dir, filename, content string) {
	err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0666)
	if err != nil {
		t.Fatalf("Failed to write %s: %v", filename, err)
	}
}
