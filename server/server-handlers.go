package server

import (
	"fmt"
	"net/http"
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
// to display the available routes. This must be the last route registered
// in order for the full set of routes to be displayed.
func (S *SimpleServer) EnableAboutHandler() {

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
			RouteDescriptor = fmt.Sprintf("%s%v", RouteDescriptor, Methods)
		}

		RouteList = append(RouteList, RouteDescriptor)
		return nil
	})

	routes := strings.Join(RouteList, "\n")
	aboutHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			w.WriteHeader(http.StatusOK)
		default:
			w.Write([]byte(routes + "\n"))
		}
	}

	S.router.HandleFunc("/about", aboutHandler)
	S.router.NotFoundHandler = http.RedirectHandler("/about", http.StatusFound)
	S.router.MethodNotAllowedHandler = http.RedirectHandler("/about", http.StatusFound)
}
