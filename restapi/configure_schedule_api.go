// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"schedule/restapi/operations"
	"schedule/restapi/operations/schedule"
)

//go:generate swagger generate server --target ..\..\schedule --name ScheduleAPI --spec ..\docs\swagger.json --principal interface{}

func configureFlags(api *operations.ScheduleAPIAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.ScheduleAPIAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	if api.ScheduleGetNextTakingHandler == nil {
		api.ScheduleGetNextTakingHandler = schedule.GetNextTakingHandlerFunc(func(params schedule.GetNextTakingParams) middleware.Responder {
			return middleware.NotImplemented("operation schedule.GetNextTaking has not yet been implemented")
		})
	}
	if api.ScheduleGetScheduleHandler == nil {
		api.ScheduleGetScheduleHandler = schedule.GetScheduleHandlerFunc(func(params schedule.GetScheduleParams) middleware.Responder {
			return middleware.NotImplemented("operation schedule.GetSchedule has not yet been implemented")
		})
	}
	if api.ScheduleGetSchedulesHandler == nil {
		api.ScheduleGetSchedulesHandler = schedule.GetSchedulesHandlerFunc(func(params schedule.GetSchedulesParams) middleware.Responder {
			return middleware.NotImplemented("operation schedule.GetSchedules has not yet been implemented")
		})
	}
	if api.SchedulePostScheduleHandler == nil {
		api.SchedulePostScheduleHandler = schedule.PostScheduleHandlerFunc(func(params schedule.PostScheduleParams) middleware.Responder {
			return middleware.NotImplemented("operation schedule.PostSchedule has not yet been implemented")
		})
	}

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
