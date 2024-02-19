package bluerpc

type Config struct {

	//Authorizer is the default authorization middleware that your app will use. Whenever you make some procedure protected this function will be used to authorize your user to proceed.
	//You will then be able to call .Auth() on your handler's context and unmarshal whatever you returned first from this function into some variable
	Authorizer *Authorizer

	//  The path where you would like the generated Typescript to be placed.
	// Keep in mind that YOU NEED TO PUT a .ts file at the end of the path
	// Default is ./output.ts
	OutputPath string

	// Boolean that determines if any typescript types will be generated.
	// Default is false. Set this to TRUE in production
	DisableGenerateTS bool

	//The function that will be used to validate your struct fields.
	ValidatorFn validatorFn

	//Disables the fiber logger middleware that is added.
	//False by default. Set this to TRUE in production
	DisableRequestLogging bool

	// by default Bluerpc transforms every error that is sent to the user into a json by the format  ErrorResponse. Turn this to true if you would like fiber to handle what form will the errors have or write your own middleware to convert all of the errors to your liking
	DisableJSONOnlyErrors bool

	//This middleware catches all of the errors of every procedure and handles responding in a proper way
	//By default it is DefaultErrorMiddleware
	ErrorMiddleware Handler

	//the URL of the server. If left empty it will be interpreted as localhost
	ServerURL string

	//disable the printing of that start server message
	DisableInfoPrinting bool

	//determines the default cors origin for every request that comes in
	CORS_Origin string

	//The address that your SSL certificate is located at
	SSLCertPath string

	//The address that your SSL key is located at
	SSLKey string

	// Puts all of the needed Pprof routes in. Read more about pprof here
	// https://pkg.go.dev/net/http/pprof
	EnablePProf bool
}

// Struct that handles all of the settings related to authorizing for routes and procedures
type Authorizer struct {
	Handler AuthHandler
}

// Creates a new authorizer struct with the defaults set
func NewAuth(handler AuthHandler) *Authorizer {
	return &Authorizer{
		Handler: handler,
	}
}
