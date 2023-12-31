package bluerpc

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/gorilla/schema"
)

// these are right from Fiber
const (
	MIMETextXML               = "text/xml"
	MIMETextHTML              = "text/html"
	MIMETextPlain             = "text/plain"
	MIMEApplicationXML        = "application/xml"
	MIMEApplicationJSON       = "application/json"
	MIMEApplicationJavaScript = "application/javascript"
	MIMEApplicationForm       = "application/x-www-form-urlencoded"
	MIMEOctetStream           = "application/octet-stream"
	MIMEMultipartForm         = "multipart/form-data"

	MIMETextXMLCharsetUTF8               = "text/xml; charset=utf-8"
	MIMETextHTMLCharsetUTF8              = "text/html; charset=utf-8"
	MIMETextPlainCharsetUTF8             = "text/plain; charset=utf-8"
	MIMEApplicationXMLCharsetUTF8        = "application/xml; charset=utf-8"
	MIMEApplicationJSONCharsetUTF8       = "application/json; charset=utf-8"
	MIMEApplicationJavaScriptCharsetUTF8 = "application/javascript; charset=utf-8"
)

type Ctx struct {
	httpR        *http.Request
	httpW        http.ResponseWriter
	indexHandler int
	nextHandler  Handler
}

// / This calls the Get method on the http Request
func (c *Ctx) Get(key string) string {
	return c.httpR.Header.Get(key)
}

// This calls the Set method on the http response
func (c *Ctx) Set(key, value string) {
	c.httpW.Header().Set(key, value)
}

// GetReqHeaders returns a map of all request headers.
func (c *Ctx) GetReqHeaders() map[string][]string {
	return c.httpR.Header
}

// GetRespHeaders returns a COPY map of all response headers.
func (c *Ctx) GetRespHeaders() map[string][]string {
	// Create a new map to hold the response headers
	respHeaders := make(map[string][]string)
	for key, values := range c.httpW.Header() {
		respHeaders[key] = make([]string, len(values))
		copy(respHeaders[key], values)
	}
	return respHeaders
}

// Slug returns the last URL segment's value (called the slug).
func (c *Ctx) Slug() (string, error) {
	parsedURL, err := url.Parse(c.httpR.URL.String())
	if err != nil {
		return "", err
	}

	// Trimming the leading slash if present
	path := strings.TrimPrefix(parsedURL.Path, "/")
	// Splitting the path by '/'
	parts := strings.Split(path, "/")

	// Getting the last part
	if len(parts) > 0 {
		return parts[len(parts)-1], nil
	}

	return "", nil
}

// Hostname returns the hostname derived from the Host HTTP header.
func (c *Ctx) Hostname() string {
	return c.httpR.Host
}

// IP returns the remote IP address of the request.
func (c *Ctx) IP() string {
	return c.httpR.RemoteAddr
}

// Links sets the Link header in the HTTP response.
func (c *Ctx) Links(link ...string) {
	if len(link)%2 != 0 {
		return // Ensure even number of parameters
	}
	var links []string
	for i := 0; i < len(link); i += 2 {
		links = append(links, `<`+link[i]+`>; rel="`+link[i+1]+`"`)
	}
	c.Set("Link", strings.Join(links, ", "))
}

// Method returns or overrides the HTTP method of the request.
func (c *Ctx) Method(override ...string) string {
	if len(override) > 0 {
		c.httpR.Method = override[0]
	}
	return c.httpR.Method
}

// Path gets or sets the path part of the request URL.
func (c *Ctx) Path(override ...string) string {
	if len(override) > 0 {
		c.httpR.URL.Path = override[0]
	}
	return c.httpR.URL.Path
}

// Protocol returns the protocol (http or https) used in the request.
func (c *Ctx) Protocol() string {
	if c.httpR.TLS != nil {
		return "https"
	}
	return "http"
}

// Attachment sets the Content-Disposition header to make the response an attachment.
// If a filename is provided, it includes it in the header.
func (c *Ctx) Attachment(filename ...string) {
	disposition := "attachment"
	if len(filename) > 0 {
		disposition += `; filename="` + filename[0] + `"`
	}
	c.Set("Content-Disposition", disposition)
}

// Decode decodes the query parameters to a struct.
//
// The first parameter must be a pointer to a struct.
func (c *Ctx) QueryParser(targetStruct interface{}) error {
	structVal := reflect.ValueOf(targetStruct).Elem()
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)
		var queryKey string

		if key := fieldType.Tag.Get("query"); key != "" {
			queryKey = key
		} else {
			queryKey = fieldType.Name
		}

		if !field.CanSet() {
			continue
		}

		queryValues, found := c.httpR.URL.Query()[queryKey]
		if !found {
			continue
		}

		switch field.Kind() {
		case reflect.Slice:
			elemKind := field.Type().Elem().Kind()
			if elemKind == reflect.Int {
				var intSlice []int
				for _, v := range queryValues {
					if intValue, err := strconv.Atoi(v); err == nil {
						intSlice = append(intSlice, intValue)
					} else {
						return fmt.Errorf("invalid integer value '%s' for query parameter '%s'", v, queryKey)
					}
				}
				field.Set(reflect.ValueOf(intSlice))
			}
			// Add cases here for slices of other types if necessary
		case reflect.String:
			field.SetString(queryValues[0])
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if intVal, err := strconv.ParseInt(queryValues[0], 10, 64); err == nil {
				field.SetInt(intVal)
			} else {
				return fmt.Errorf("invalid integer value '%s' for query parameter '%s'", queryValues[0], queryKey)
			}
			// Add more cases as needed for other types
		}
	}

	return nil
}
func setField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	fieldVal := structValue.FieldByName(name)

	if !fieldVal.IsValid() {
		return fmt.Errorf("invalid query parameter: '%s' is not a recognized field", name)
	}

	if !fieldVal.CanSet() {
		return fmt.Errorf("unable to set query parameter: '%s' due to field restrictions", name)
	}

	// Convert value to the correct type
	val := reflect.ValueOf(value)
	fieldType := fieldVal.Type()
	if val.Type().ConvertibleTo(fieldType) {
		fieldVal.Set(val.Convert(fieldType))
	} else {
		return fmt.Errorf("type mismatch for query parameter '%s': expected %s, got %s", name, fieldType, val.Type())
	}

	return nil
}
func (c *Ctx) BodyParser(targetStruct interface{}) error {
	contentType := c.httpR.Header.Get("Content-Type")

	if contentType == "" {
		contentType = "application/json"
	}

	switch {
	case strings.Contains(contentType, "application/json"):
		return c.decodeJSON(targetStruct)
	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		return c.decodeForm(targetStruct)
	//TODO
	default:
		return http.ErrNotSupported
	}
}

func (c *Ctx) decodeJSON(targetStruct interface{}) error {
	if c.httpR.ContentLength == 0 {
		return fmt.Errorf("The body of the request is empty")
	}
	decoder := json.NewDecoder(c.httpR.Body)
	return decoder.Decode(targetStruct)
}

func (c *Ctx) decodeForm(targetStruct interface{}) error {
	if err := c.httpR.ParseForm(); err != nil {
		return err
	}
	return schema.NewDecoder().Decode(targetStruct, c.httpR.Form)
}

func (c *Ctx) Cookie(cookie *http.Cookie) {
	http.SetCookie(c.httpW, cookie)
}

type Map map[string]interface{}

func (c *Ctx) JSON(data interface{}) error {

	// Marshal the struct into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err

	}
	c.httpW.Header().Set("Content-Type", "application/json")
	c.httpW.Write(jsonData)
	return nil

}
func (c *Ctx) Status(code int) *Ctx {
	c.httpW.WriteHeader(code)
	return c
}

func (c *Ctx) SendString(str string) error {
	c.httpW.Header().Set("Content-Type", "text/plain")

	_, err := c.httpW.Write([]byte(str))
	if err != nil {
		return err
	}

	return nil
}

func (c *Ctx) XML(data interface{}) error {
	xmlData, err := xml.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	c.httpW.Header().Set("Content-Type", "application/xml")
	c.httpW.Write(xmlData)
	return nil
}
