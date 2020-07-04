package main

// Plugins must supply their own middlewares if they want to have additional
// specific logic applied to the Handlers it returns. These middleware(s)
// will only be applied to the handlers and routes presented by the plugin
// itself, if you wish to have middlewares applies to the entire tree, or
// shared by multiple modules, this logic will have to be injected somewhere
// else.
//
// The middleware must return an http.Handler, however this doesn't
// mean that these middlewares can't operate on functions that don't implement
// the http.Handler interface (http.HandlerFunc). This only means that these
// middleware functions have to form a closure over the more complex function,
// such that it returns a compatible http.Handler.
//
// For example:
//
// func nameMiddleware() http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//
// 		Name := mux.Vars(r)["Name"]
// 		if Name == "" {
// 			ExitHandler(w, http.StatusBadRequest, "No [ Name ] parameter found in URL", errors.New("request error: Missing URL request parameter"))
// 			return
// 		}
//
// 		echoHandler(w, r, Name)
// 	})
// }
//
// func echoHandler(w http.ResponseWriter, r *http.Request, Name string) {
// 	w.Write([]byte(Name))
// }
