package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

// SimpleHandler represents a simplification to the standard http handlerFuncs,
// allowing simpler registration and logging with Routers.
type SimpleHandler struct {
	Handler http.HandlerFunc
	Path    string
	Methods []string
}

// NewSimpleHandler will create and return a new SimpleHandler, ready to be used.
func NewSimpleHandler(h http.HandlerFunc, Path string, Methods ...string) SimpleHandler {

	if !strings.HasPrefix(Path, "/") {
		Path = "/" + Path
	}

	return SimpleHandler{
		Handler: h,
		Path:    Path,
		Methods: Methods,
	}
}

// notFoundHandler represents the default function to call if the path
// requested has not been registered.
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

// methodNotAllowedHandler represents the default function to call if
// the method used on the path has not been registered explicitly.
func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// EnableAboutHandler will enable and set up the "about" handler,
// to display the available routes at "/about". This will only be able to
// display information about routes registered before this function is called.
// If additional routes are registered after this is called, they will not be
// displayed unless this is called again.
//
// This function will cause the NotFound and MethodNotAllowed
// handlers to redirect to the /about page.
func (S *SimpleServer) EnableAboutHandler() {

	// Walk the router, printing out an API summary for each route in the order the router
	// will be searched.
	RouteList := []string{}
	S.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		RouteDescriptor := ""

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
			RouteDescriptor = fmt.Sprintf("%s%v ", RouteDescriptor, Methods)
		}

		RouteList = append(RouteList, RouteDescriptor)
		return nil
	})

	// Format the route information one node per row
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
	S.Router().NotFoundHandler = http.RedirectHandler("/about", http.StatusFound)
	S.Router().MethodNotAllowedHandler = http.RedirectHandler("/about", http.StatusFound)
}

// RegisterSPAHandler will regitster an HTTP Handler to allow serving a Single Page Application.
// The application will be based off URLBase, and will serve content based out of PathBase.
//
// The URLBase must be the same as what's defined to be <base href="/URLBase"> within
// the SPA.
//
// This function will also set up the NotFound handler to redirect to URLBase/index.html.
//
// This handler is fully able to be served on the same server as the raw API nodes,
// as long as the URLBase path is a distinct URL tree.
func (S *SimpleServer) RegisterSPAHandler(URLBase, PathBase string) error {

	// Assert that the URLBase exists
	if URLBase == "" {
		URLBase = "/"
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
	err = S.router.PathPrefix(URLBase).Handler(http.StripPrefix(URLBase, http.FileServer(http.Dir(AbsPathBase)))).GetError()

	// Redirect any not-found routes to redirect to the main page.
	// This will override any existing NotFoundHandler, so if this functionality is NOT desired,
	// S.Router().NotFoundHandler will have to be assigned after this function
	S.Router().NotFoundHandler = http.RedirectHandler(URLBase+"/index.html", http.StatusPermanentRedirect)

	return err
}
