// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"

	"db-forum/restapi/operations"
)

//go:generate swagger generate server --target .. --name bd-forum-gen --spec ../swagger-8.yaml

func configureFlags(api *operations.BdForumGenAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.BdForumGenAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.BinConsumer = runtime.ByteStreamConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.ClearHandler = operations.ClearHandlerFunc(func(params operations.ClearParams) middleware.Responder {
		return middleware.NotImplemented("operation .Clear has not yet been implemented")
	})
	api.ForumCreateHandler = operations.ForumCreateHandlerFunc(func(params operations.ForumCreateParams) middleware.Responder {
		return middleware.NotImplemented("operation .ForumCreate has not yet been implemented")
	})
	api.ForumGetOneHandler = operations.ForumGetOneHandlerFunc(func(params operations.ForumGetOneParams) middleware.Responder {
		return middleware.NotImplemented("operation .ForumGetOne has not yet been implemented")
	})
	api.ForumGetThreadsHandler = operations.ForumGetThreadsHandlerFunc(func(params operations.ForumGetThreadsParams) middleware.Responder {
		return middleware.NotImplemented("operation .ForumGetThreads has not yet been implemented")
	})
	api.ForumGetUsersHandler = operations.ForumGetUsersHandlerFunc(func(params operations.ForumGetUsersParams) middleware.Responder {
		return middleware.NotImplemented("operation .ForumGetUsers has not yet been implemented")
	})
	api.PostGetOneHandler = operations.PostGetOneHandlerFunc(func(params operations.PostGetOneParams) middleware.Responder {
		return middleware.NotImplemented("operation .PostGetOne has not yet been implemented")
	})
	api.PostUpdateHandler = operations.PostUpdateHandlerFunc(func(params operations.PostUpdateParams) middleware.Responder {
		return middleware.NotImplemented("operation .PostUpdate has not yet been implemented")
	})
	api.PostsCreateHandler = operations.PostsCreateHandlerFunc(func(params operations.PostsCreateParams) middleware.Responder {
		return middleware.NotImplemented("operation .PostsCreate has not yet been implemented")
	})
	api.StatusHandler = operations.StatusHandlerFunc(func(params operations.StatusParams) middleware.Responder {
		return middleware.NotImplemented("operation .Status has not yet been implemented")
	})
	api.ThreadCreateHandler = operations.ThreadCreateHandlerFunc(func(params operations.ThreadCreateParams) middleware.Responder {
		return middleware.NotImplemented("operation .ThreadCreate has not yet been implemented")
	})
	api.ThreadGetOneHandler = operations.ThreadGetOneHandlerFunc(func(params operations.ThreadGetOneParams) middleware.Responder {
		return middleware.NotImplemented("operation .ThreadGetOne has not yet been implemented")
	})
	api.ThreadGetPostsHandler = operations.ThreadGetPostsHandlerFunc(func(params operations.ThreadGetPostsParams) middleware.Responder {
		return middleware.NotImplemented("operation .ThreadGetPosts has not yet been implemented")
	})
	api.ThreadUpdateHandler = operations.ThreadUpdateHandlerFunc(func(params operations.ThreadUpdateParams) middleware.Responder {
		return middleware.NotImplemented("operation .ThreadUpdate has not yet been implemented")
	})
	api.ThreadVoteHandler = operations.ThreadVoteHandlerFunc(func(params operations.ThreadVoteParams) middleware.Responder {
		return middleware.NotImplemented("operation .ThreadVote has not yet been implemented")
	})
	api.UserCreateHandler = operations.UserCreateHandlerFunc(func(params operations.UserCreateParams) middleware.Responder {
		return middleware.NotImplemented("operation .UserCreate has not yet been implemented")
	})
	api.UserGetOneHandler = operations.UserGetOneHandlerFunc(func(params operations.UserGetOneParams) middleware.Responder {
		return middleware.NotImplemented("operation .UserGetOne has not yet been implemented")
	})
	api.UserUpdateHandler = operations.UserUpdateHandlerFunc(func(params operations.UserUpdateParams) middleware.Responder {
		return middleware.NotImplemented("operation .UserUpdate has not yet been implemented")
	})

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
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
