package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// QueryKeyValue is a key-value pair used to register routes with URL Query parameters.
type QueryKeyValue struct {
	Key   string
	Value string
}

// SimpleHandler represents a simplification to the standard http handlerFuncs,
// allowing simpler registration and logging with Routers.
type SimpleHandler struct {

	// The actual HTTP Handler to call when this is matched
	Handler http.Handler `json:"-"`

	// Optional: To allow for a route to only match on a specific host or host:port
	Host string `json:",omitempty"`

	// Required: The URL Path to match the handler on
	Path string `json:",omitempty"`

	// Optional: The set of methods this handler will be operated for
	Methods []string `json:",omitempty"`

	// Optional: The set of Key/Value URL Query strings which must match
	// in order for this handler to be called.
	// See the mux.Route.Queries() for more details on the form and shape
	// of these values.
	Queries []QueryKeyValue `json:",omitempty"`

	// Optional: An additional description of the route, to provide additional context
	// and understanding when displayed via the "/about" handler.
	Description string `json:",omitempty"`
}

// NewSimpleHandler will create and return a new SimpleHandler, ready to be used.
// This constructor does not include the ability to specify a host or query parameters
// both to keep the argument list small, and because these are substantially less common
// than the URL path and acceptable methods. See the SimpleHandler.AddQueries()
// and SimpleHandler.AddHost() functions to specify these properties.
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

// AddQueries will add the Key/Value pairs to the Handler to be included when registered with the server.
// See the rules for mux.Route.Queries() for more details on the shape of these QueryPairs.
func (H *SimpleHandler) AddQueries(QueryPairs ...string) error {

	// Assert there's an even number of values given, to build the proper pairs
	if len(QueryPairs)%2 != 0 {
		return errors.New("easytls handler error: Odd number of query-strings provided")
	}

	// Walk the given set of Key/Value pairs, adding them to the handler.
	for i := 0; i < len(QueryPairs); i += 2 {
		H.Queries = append(H.Queries, QueryKeyValue{QueryPairs[i], QueryPairs[i+1]})
	}

	return nil
}

// AddHost will configure the handler to only match for a given host or host:port argument.
// See the rules for mux.Route.Host() for more details on the shape of the Host string.
func (H *SimpleHandler) AddHost(Host string) {
	H.Host = Host
}

// AddDescription will add an optional description field to the handler, which will be shown alongside
// the match criteria on the "/about" handler.
func (H *SimpleHandler) AddDescription(Description string) {
	H.Description = Description
}

// AddPrefixToRoutes will assert that all routes have a given prefix
func AddPrefixToRoutes(Prefix string, Handlers ...SimpleHandler) []SimpleHandler {

	if Prefix == "" {
		Prefix = "/"
	}

	if !strings.HasPrefix(Prefix, "/") {
		Prefix = "/" + Prefix
	}

	if !strings.HasSuffix(Prefix, "/") {
		Prefix = Prefix + "/"
	}

	for i := range Handlers {
		Handlers[i].Path = strings.ReplaceAll(Prefix+Handlers[i].Path, "//", "/")
	}

	return Handlers
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

	aboutHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			w.WriteHeader(http.StatusOK)
		default:
			enc := json.NewEncoder(w)
			enc.SetIndent("", "\t")
			enc.SetEscapeHTML(true)
			enc.Encode(S.routes)
		}
	}

	S.Router().Path("/about").HandlerFunc(aboutHandler)
}

// RegisterSPAHandler will register an HTTP Handler to allow serving a Single Page Application.
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
	H := NewSimpleHandler(http.StripPrefix(URLBase, http.FileServer(http.Dir(AbsPathBase))), URLBase)
	H.AddDescription(fmt.Sprintf("Serve the Single Page Web Application from [ %s ] at [ %s ]", PathBase, URLBase))
	S.AddSubrouter(S.Router(), URLBase, H)

	return nil
}
