package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gorilla/mux"
)

// SimpleHandler represents a simplification to the standard http handlerFuncs,
// allowing simpler registration and logging with Routers.
type SimpleHandler struct {
	Handler http.Handler
	Path    string
	Methods []string
}

// NewSimpleHandler will create and return a new SimpleHandler, ready to be used.
func NewSimpleHandler(h http.Handler, Path string, Methods ...string) SimpleHandler {

	if !strings.HasPrefix(Path, "/") {
		Path = "/" + Path
	}

	return SimpleHandler{
		Handler: h,
		Path:    Path,
		Methods: Methods,
	}
}

// NotFoundHandler represents the default handler to use for a route that doesn't exist,
// or used as a mechanism to "remove" a route by replacing the existing http.Handler
// with this.
func NotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
}

// MethodNotAllowedHandler represents the default function to call if
// the method used on the path has not been registered explicitly.
func MethodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}

// EnableAboutHandler will enable and set up the "about" handler,
// to display the available routes at "/about". This will only be able to
// display information about routes registered before this function is called.
// If additional routes are registered after this is called, they will not be
// displayed unless this is called again.
func (S *SimpleServer) enableAboutHandler() {

	// Walk the router, printing out an API summary for each route in the order the router
	// will be searched.
	RouteList := []string{}
	S.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		RouteDescriptor := ""

		if route.GetHandler() == nil {
			return nil
		}

		if Host, err := route.GetHostTemplate(); err == nil {
			RouteDescriptor = fmt.Sprintf("%s%s ", RouteDescriptor, Host)
		}

		if Path, err := route.GetPathTemplate(); err == nil {
			RouteDescriptor = fmt.Sprintf("%s%s ", RouteDescriptor, Path)
		}

		if Queries, err := route.GetQueriesTemplates(); err == nil && len(Queries) > 0 {
			RouteDescriptor = fmt.Sprintf("%s%v ", RouteDescriptor, Queries)
		}

		if Methods, err := route.GetMethods(); err == nil && len(Methods) > 0 {
			sort.Strings(Methods)
			RouteDescriptor = fmt.Sprintf("%s%v ", RouteDescriptor, Methods)
		}

		RouteList = append(RouteList, RouteDescriptor)
		return nil
	})

	// Format the route information one node per row
	sort.Strings(RouteList)
	routes := strings.Join(RouteList, "\n")
	aboutHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			w.WriteHeader(http.StatusOK)
		default:
			w.Write([]byte(routes + "\n"))
		}
	}

	S.Router().Path("/about").HandlerFunc(aboutHandler)
}

// RegisterSPAHandler will regitster an HTTP Handler to allow serving a Single Page Application.
// The application will be based off URLBase, and will serve content based out of PathBase.
//
// The URLBase must be the same as what's defined to be <base href="/URLBase"> within
// the SPA.
//
// This handler is fully able to be served on the same server as the raw API nodes,
// as long as the URLBase path is a distinct URL tree.
func (S *SimpleServer) RegisterSPAHandler(URLBase, PathBase string) error {

	// Assert that the PathBase is formatted correctly
	if !strings.HasSuffix(PathBase, "/") {
		PathBase += "/"
	}

	// Assert that the PathBase to serve from exists
	AbsPathBase, err := filepath.Abs(PathBase)
	if err != nil {
		return err
	}
	if _, err := os.Stat(AbsPathBase); err != nil {
		return err
	}

	// Register a route to handle anything with the URLBase prefix
	// Then strip the prefix and expose a fileserver on the directory.
	S.AddSubrouter(S.Router(), URLBase, NewSimpleHandler(http.StripPrefix(URLBase, http.FileServer(http.Dir(AbsPathBase))), URLBase))

	return nil
}
