package bluerpc

type RuntimeEnv string

const (
	DEVELOPMENT RuntimeEnv = "development"
	PRODUCTION  RuntimeEnv = "production"
)

type Config struct {
	//  The path where you would like the generated Typescript to be placed.
	// Default is ./output.ts
	OutputPath string

	// Boolean that determines if any typescript types will be generated.
	// Default is false. Set this to TRUE in production
	disableGenerateTS bool

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
}
