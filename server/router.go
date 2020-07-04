package server

import (
	"fmt"
	"strings"

	"github.com/gorilla/mux"
)

// NewDefaultRouter will create a new Router, based on the gorilla/mux
// implementation. This will pre-set the "trailing-slash" behaviour to not be
// pedantic, as well as initialing the "Not-Found" and "Method-Not-Allowed"
// behaviours to simply return the corresponding status codes.
// These can be overridden if desired.
func NewDefaultRouter() *mux.Router {
	r := mux.NewRouter()

	// Don't be pedantic about possible trailing slashes in the routes.
	r.StrictSlash(true)

	// Register some default handlers
	r.NotFoundHandler = NotFoundHandler()
	r.MethodNotAllowedHandler = MethodNotAllowedHandler()

	return r
}

// AddMiddlewares is a convenience wrapper for the mux.Router "Use" function.
// This will add the middlewares to the router in the order specified,
// which also defines their execution order.
func (S *SimpleServer) AddMiddlewares(middlewares ...MiddlewareHandler) {
	for _, mwf := range middlewares {
		S.router.Use(mux.MiddlewareFunc(mwf))
	}
}

// AddHandlers will add the given handlers to the router, with the verbose
// flag determining if a log message should be generated for each added route.
func (S *SimpleServer) AddHandlers(Router *mux.Router, Handlers ...SimpleHandler) {
	S.addHandlers(Router, Handlers...)
}

// AddSubrouter will add the set of Handlers to the server by creating a dedicated Subrouter
// of the given Router.
//
// Currently, these subrouters are defined based on PathPrefix.
func (S *SimpleServer) AddSubrouter(Router *mux.Router, PathPrefix string, Handlers ...SimpleHandler) {

	if !strings.HasSuffix(PathPrefix, "/") {
		PathPrefix += "/"
	}

	// Don't create subrouters for URLRoots
	if PathPrefix == "/" {
		S.AddHandlers(Router, Handlers...)
	} else {
		s := Router.PathPrefix(PathPrefix).Subrouter()
		S.Logger().Printf("Creating subrouter for PathPrefix [ %s ] on Server at [ %s ]", PathPrefix, S.Addr())
		S.addHandlers(s, Handlers...)
	}
}

func (S *SimpleServer) addHandlers(Router *mux.Router, Handlers ...SimpleHandler) {
	// Register the routes, this IS order dependent.
	for _, Node := range Handlers {

		RouteDescriptor := ""

		// Create a route for the handler
		Route := S.router.NewRoute().Handler(Node.Handler)

		// Assign the path
		if Node.Path != "" {
			Route = Route.PathPrefix(Node.Path)
			RouteDescriptor = fmt.Sprintf("%s[ %s ] ", RouteDescriptor, Node.Path)
		}

		// Assign any methods
		if len(Node.Methods) != 0 {
			Route = Route.Methods(Node.Methods...)
			RouteDescriptor = fmt.Sprintf("%s%v ", RouteDescriptor, Node.Methods)
		}

		// ...

		S.Logger().Printf("Added route: %sto server at [ %s ]", RouteDescriptor, S.Addr())
	}
}
