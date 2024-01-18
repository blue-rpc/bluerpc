package bluerpc

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/gorilla/schema"
)

type Ctx struct {
	httpR       *http.Request
	httpW       http.ResponseWriter
	nextHandler Handler
}

// / This calls the Get method on the http Request to get a value from the header depending on a given key
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

//	decodes the query parameters to a struct.
//
// The first parameter must be a pointer to a struct.
func (c *Ctx) queryParser(targetStruct interface{}, path string) error {
	structVal := reflect.ValueOf(targetStruct).Elem()
	structType := structVal.Type()
	query := c.httpR.URL.Query()

	// I take the slug key and split it to get all of the possible nested routes that the user might add
	// for example /:userId/:imageId/name will be split into [":userId". ":imageId", "name"]
	slugKeys := strings.Split(path, "/")
	slices.Reverse(slugKeys)

	//I only care about the slugs if they are dynamic. yet I also care about their position in the url, for example I need to know that /:userId is second to last in /:userId/:imageId/name
	// so if it is not dynamic I leave it blank and if it is I remove the dot to be able to compare it later
	for i, slugKey := range slugKeys {
		if !strings.HasPrefix(slugKey, ":") {
			slugKeys[i] = ""
		} else {
			slugKeys[i] = strings.TrimPrefix(slugKey, ":")
		}
	}

	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)
		var queryKey string

		if key := fieldType.Tag.Get("paramName"); key != "" {
			queryKey = key
		} else {
			queryKey = fieldType.Name
		}

		if !field.CanSet() {
			continue
		}

		var queryValues []string

		//continuing from the last comment
		// if the struct query key matches any
		if posOfSlugInUrl := findIndex(slugKeys, queryKey); posOfSlugInUrl != -1 {
			url := c.httpR.URL.Path
			urlParts := strings.Split(url, "/")

			queryValues = append(queryValues, urlParts[len(urlParts)-posOfSlugInUrl-1])
		} else {
			vals, found := query[queryKey]
			if !found {
				continue
			}
			queryValues = append(queryValues, vals...)

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
		case reflect.Bool:
			if boolVal, err := strconv.ParseBool(queryValues[0]); err == nil {
				field.SetBool(boolVal)
			} else {
				return fmt.Errorf("invalid boolean value '%s' for query parameter '%s'", queryValues[0], path)
			}
		default:
			return fmt.Errorf("unsupported type '%s' for query parameter '%s'", field.Kind(), path)

		}
	}

	return nil
}

func (c *Ctx) bodyParser(targetStruct interface{}) error {
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
	var dataMap map[string]interface{}
	body, err := io.ReadAll(c.httpR.Body)

	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &dataMap); err != nil {
		return err
	}

	// Iterate over the fields of the struct
	val := reflect.ValueOf(targetStruct).Elem()
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldInfo := typ.Field(i)

		// Determine the key name to look for in the JSON
		keyName := fieldInfo.Tag.Get("paramName")
		jsonName := fieldInfo.Tag.Get("JSON")
		if keyName == "" {
			if jsonName != "" {
				keyName = jsonName
			} else {
				keyName = fieldInfo.Name // Fallback to field's name
			}
		}

		// Use strings.ToLower or another strategy if JSON keys are case-insensitive
		keyName = strings.ToLower(keyName)

		// Assign the corresponding value from the map
		if jsonValue, ok := dataMap[keyName]; ok {
			fieldValue := reflect.ValueOf(jsonValue)
			if field.Type() == fieldValue.Type() {
				field.Set(fieldValue)
			} else {
				// Convert and set value if types are different
				// This part might need more robust error handling and type conversion
				convertedValue := reflect.ValueOf(jsonValue).Convert(field.Type())
				field.Set(convertedValue)
			}
		}
	}

	return nil
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

func (c *Ctx) jSON(data interface{}) error {

	// Marshal the struct into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err

	}
	c.httpW.Header().Set("Content-Type", "application/json")
	c.httpW.Write(jsonData)
	return nil

}
func (c *Ctx) status(code int) *Ctx {
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

func (c *Ctx) xML(data interface{}) error {
	xmlData, err := xml.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	c.httpW.Header().Set("Content-Type", "application/xml")
	c.httpW.Write(xmlData)
	return nil
}
