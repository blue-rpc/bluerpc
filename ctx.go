package bluerpc

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gorilla/schema"
)

type Ctx struct {
	httpR       *http.Request
	httpW       http.ResponseWriter
	nextHandler Handler
	auth        any

	// This session field can be used in your middlewares for you to store any data that you would need to pass on to your handlers
	Session any
}

// Gets the authorization set data by your authorization function on the context
// If there is no data OR if the data you've set on your context does not match the generic type you provided then it will trigger a panic
func GetAuth[authType any](ctx *Ctx) authType {
	if ctx.auth == nil {
		panic("There is no context auth to provide. You did not make this route protected in order to call the Auth() function on. You can only call the Auth() function on protected procedures")
	}
	castedAuth, ok := ctx.auth.(authType)
	if ok {
		return castedAuth
	} else {
		panic("Your provided generic argument does not match the type that you've returned from your authorization function")
	}
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
func (c *Ctx) queryParser(targetStruct interface{}, slug string) error {
	v := reflect.ValueOf(targetStruct)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("targetStruct must be a non-nil pointer")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("targetStruct must point to a struct")
	}

	// Extract query parameters
	query := c.httpR.URL.Query()

	// I take the slug key and split it to get all of the possible nested routes that the user might add
	// for example /:userId/:imageId/name will be split into [":userId". ":imageId", "name"]
	slugKeys := strings.Split(slug, "/")
	lenSlugKeys := len(slugKeys)
	url := c.httpR.URL.Path
	urlParts := strings.Split(url, "/")
	lenUrlParts := len(urlParts)

	//should never happen
	if lenUrlParts < lenSlugKeys {
		panic("The url parts are (for some very unexpected reason) shorter than the slug parts")
	}

	urlPartsOfTheSlug := urlParts[lenUrlParts-lenSlugKeys:]

	//I only care about the slugs if they are dynamic. yet I also care about their position in the url, for example I need to know that /:userId is second to last in /:userId/:imageId/name
	// the position in the array determines the position of the slug in the url
	//so I will leave the non-dynamic slug elements as filler "" in the arrays. If I then find something that matches the slug then it should work
	for i, slugKey := range slugKeys {
		if !strings.HasPrefix(slugKey, ":") {
			slugKeys[i] = ""
		} else {
			slugKeys[i] = strings.TrimPrefix(slugKey, ":")
		}
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue // Skip unexported fields
		}

		fieldType := v.Type().Field(i)
		queryKey := fieldType.Tag.Get("paramName")
		if queryKey == "" {
			queryKey = fieldType.Name
		}

		// Check if the field corresponds to a URL slug
		var values []string
		posOfSlugInUrl := findIndex(slugKeys, queryKey)

		if posOfSlugInUrl != -1 {
			values = append(values, urlPartsOfTheSlug[posOfSlugInUrl])
		}

		// If not a slug or no value found, look for a query parameter
		if len(values) == 0 {
			queryValues, ok := query[queryKey]
			if !ok {
				pathValue := c.httpR.PathValue(queryKey)
				if len(pathValue) != 0 {
					values = append(values, pathValue)
				} else {
					continue
				}
			}
			for _, value := range queryValues {
				splitQueryValues := strings.Split(value, ",")
				values = append(values, splitQueryValues...)
			}
		}

		// If values are found, attempt to set the field
		if err := setField(field, values); err != nil {
			return fmt.Errorf("failed to set field '%s': %v", fieldType.Name, err)
		}
	}

	return nil
}

// returns the byte size depending on the number kind. panics if the passed variable is not a kind

func (c *Ctx) bodyParser(targetStruct interface{}) error {
	contentType := c.httpR.Header.Get("Content-Type")

	if contentType == "" {
		contentType = "application/json"
	}

	switch {
	case strings.Contains(contentType, TextPlain):
		// THIS MIGHT BE AN ISSUE LATER OR LEFT TO DO. NOW IT ASSUMES TEXT/PLAIN IS JUST JSON
		return c.decodeJSON(targetStruct)
	case strings.Contains(contentType, ApplicationJSON):
		return c.decodeJSON(targetStruct)
	case strings.Contains(contentType, ApplicationForm):
		return c.decodeForm(targetStruct)
	//TODO
	default:
		return http.ErrNotSupported
	}
}

func (c *Ctx) decodeJSON(target interface{}) error {
	// Ensure target is a pointer to a struct
	if reflect.TypeOf(target).Kind() != reflect.Ptr || reflect.TypeOf(target).Elem().Kind() != reflect.Struct {
		return errors.New("target must be a pointer to a struct")
	}

	// Decode the JSON body into a map
	var dataMap map[string]interface{}
	if err := json.NewDecoder(c.httpR.Body).Decode(&dataMap); err != nil {
		return err
	}

	// Get the reflect value and type of the target struct
	val := reflect.ValueOf(target).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Use paramName tag as priority, fallback to field name
		paramName := fieldType.Tag.Get("paramName")
		if paramName == "" {
			paramName = strings.ToLower(fieldType.Name)
		}

		// Find the corresponding JSON value
		jsonValue, exists := dataMap[paramName]
		if !exists {
			continue
		}

		// Ensure the field can be set
		if !field.CanSet() {
			continue
		}

		// Convert and set the field value
		if err := setFieldValue(field, jsonValue); err != nil {
			return fmt.Errorf("failed to set field '%s': %v", fieldType.Name, err)
		}
	}

	return nil
}
func setFieldValue(field reflect.Value, value interface{}) error {
	if value == nil {
		return nil
	}

	// Convert value to the field's type and set it
	valueType := reflect.TypeOf(value)
	fieldType := field.Type()

	if valueType.AssignableTo(fieldType) {
		field.Set(reflect.ValueOf(value))
		return nil
	} else if valueType.ConvertibleTo(fieldType) {
		field.Set(reflect.ValueOf(value).Convert(fieldType))
		return nil
	}

	return fmt.Errorf("type mismatch: cannot assign %v to %v", valueType, fieldType)
}
func (c *Ctx) decodeForm(targetStruct interface{}) error {
	if err := c.httpR.ParseForm(); err != nil {
		return err
	}
	return schema.NewDecoder().Decode(targetStruct, c.httpR.Form)
}

// returns an array of all of the cookies on the REQUEST object
func (c *Ctx) GetCookies() []*http.Cookie {
	return c.httpR.Cookies()
}

// returns a specific cookie given the string
func (c *Ctx) GetCookie(key string) (*http.Cookie, error) {
	return c.httpR.Cookie(key)
}
func (c *Ctx) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.httpW, cookie)
}

type Map map[string]interface{}

func (c *Ctx) jSON(data interface{}) error {
	jsonData, err := c.marshalJSON(data)
	if err != nil {
		return err
	}

	c.httpW.Header().Set("Content-Type", "application/json")
	c.httpW.Write(jsonData)
	return nil
}

func (c *Ctx) marshalJSON(data interface{}) ([]byte, error) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		result := make(map[string]interface{})
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			key, fieldValue := c.getFieldKeyAndValue(field, v.Field(i))
			if !fieldValue.IsValid() || !field.IsExported() {
				continue
			}
			switch fieldValue.Kind() {
			case reflect.Struct, reflect.Slice, reflect.Array, reflect.Map:
				if fieldValue.Kind() == reflect.Map && fieldValue.Type().Key().Kind() != reflect.String {
					result[key] = fieldValue.Interface()
					continue
				}
				// Recursively process nested structs, slices, arrays, and string-keyed maps
				subResult, err := c.marshalJSON(fieldValue.Interface())
				if err != nil {
					return nil, err
				}
				// For slices and arrays, unmarshal the result to maintain type
				if fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Array {
					var sliceResult []interface{}
					if err := json.Unmarshal(subResult, &sliceResult); err != nil {
						return nil, err
					}
					result[key] = sliceResult
				} else {
					var mapOrStructResult interface{}
					if err := json.Unmarshal(subResult, &mapOrStructResult); err != nil {
						return nil, err
					}
					result[key] = mapOrStructResult
				}
			default:
				result[key] = fieldValue.Interface()
			}
		}
		return json.Marshal(result)
	case reflect.Slice, reflect.Array:
		// Handle slices and arrays at the top level
		sliceResult := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			processedElem, err := c.marshalJSON(elem.Interface())
			if err != nil {
				return nil, err
			}
			var intermediate interface{}
			if err := json.Unmarshal(processedElem, &intermediate); err != nil {
				return nil, err
			}
			sliceResult[i] = intermediate
		}
		return json.Marshal(sliceResult)
	default:
		return json.Marshal(data)
	}
}

func (c *Ctx) getFieldKeyAndValue(field reflect.StructField, value reflect.Value) (string, reflect.Value) {
	jsonTag := field.Tag.Get("json")
	paramNameTag := field.Tag.Get("paramName")

	key := field.Name
	if jsonTag != "" {
		key = jsonTag
	} else if paramNameTag != "" {
		key = paramNameTag
	}
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	key = strings.Split(key, ",")[0] // Remove options like omitempty
	return key, value
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
