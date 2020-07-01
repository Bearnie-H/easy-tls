package plugins

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/Bearnie-H/easy-tls/server"
	"github.com/gorilla/mux"
)

// URL Definitions, to easily share between client and server sides of the channel
const (
	URLStartHandler   string = "/start/{Name}"
	URLRestartHandler string = "/restart/{Name}"
	URLReloadHandler  string = "/reload/{Name}"
	URLVersionHandler string = "/version/{Name}"
	URLStopHandler    string = "/stop/{Name}"
	URLStateHandler   string = "/state/{Name}"
	URLActiveHandler  string = "/active"
	URLHelpHandler    string = "/"
)

var availableOptions = []string{
	"start {name...}",
	"restart {name...}",
	"reload {name...}",
	"version {name...}",
	"state {name...}",
	"stop {name...}",
	"active",
	"help",
}

// FormatCommandHandlers will format and prepare the full set of handlers used by the
// Plugin Command Server
func formatCommandHandlers(Agent *Agent) []server.SimpleHandler {

	h := []server.SimpleHandler{}
	h = append(h, server.NewSimpleHandler(startHandler(Agent), URLStartHandler, http.MethodGet, http.MethodPost))
	h = append(h, server.NewSimpleHandler(restartHandler(Agent), URLRestartHandler, http.MethodGet, http.MethodPost))
	h = append(h, server.NewSimpleHandler(reloadHandler(Agent), URLReloadHandler, http.MethodGet, http.MethodPost))
	h = append(h, server.NewSimpleHandler(versionHandler(Agent), URLVersionHandler, http.MethodGet, http.MethodPost))
	h = append(h, server.NewSimpleHandler(stopHandler(Agent), URLStopHandler, http.MethodGet, http.MethodPost))
	h = append(h, server.NewSimpleHandler(stateHandler(Agent), URLStateHandler, http.MethodGet, http.MethodPost))
	h = append(h, server.NewSimpleHandler(activeHandler(Agent), URLActiveHandler, http.MethodGet, http.MethodPost))
	h = append(h, server.NewSimpleHandler(helpHandler(Agent), URLHelpHandler, http.MethodGet, http.MethodPost))

	return h
}

// ExitHandler standardizes all exit points of these handlers to use the same formatting
func exitHandler(w http.ResponseWriter, StatusCode int, Message string, Err error, args ...interface{}) PluginStatus {
	w.WriteHeader(StatusCode)
	s := NewPluginStatus(fmt.Sprintf(Message, args...), Err, false)
	w.Write([]byte(s.String()))
	return s
}

// HelloHandler is a test function, used to see if the server is working
func helpHandler(Agent *Agent) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := bytes.NewBuffer(nil)

		b.WriteString("Plugin Agent Options:")
		for _, o := range availableOptions {
			b.WriteString(fmt.Sprintf("\n\t%s", o))
		}

		w.WriteHeader(http.StatusOK)
		w.Write(b.Bytes())
	})
}

// StartHandler will accept a request and attempt to
func startHandler(Agent *Agent) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		Name := mux.Vars(r)["Name"]

		M, err := Agent.GetByName(Name)
		if err != nil {
			s := exitHandler(w, http.StatusNotFound, "No module matching name [ %s ] could be found", err, Name)
			Agent.Logger().Println(s.String())
			return
		}

		if err := M.Start(); err != nil {
			s := exitHandler(w, http.StatusInternalServerError, "Failed to start module [ %s ]", err, M.Name())
			Agent.Logger().Println(s.String())
		} else {
			s := exitHandler(w, http.StatusOK, "Successfully started module [ %s ]", nil, M.Name())
			Agent.Logger().Println(s.String())
		}
	})
}

// RestartHandler will accept a request and attempt to
func restartHandler(Agent *Agent) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		Name := mux.Vars(r)["Name"]

		M, err := Agent.GetByName(Name)
		if err != nil {
			s := exitHandler(w, http.StatusNotFound, "No module matching name [ %s ] could be found", err, Name)
			Agent.Logger().Println(s.String())
			return
		}

		if err := M.Stop(); err != nil {
			s := exitHandler(w, http.StatusInternalServerError, "Failed to stop module [ %s ]", err, M.Name())
			Agent.Logger().Println(s.String())
			return
		}

		if err := M.Start(); err != nil {
			s := exitHandler(w, http.StatusInternalServerError, "Failed to start module [ %s ]", err, M.Name())
			Agent.Logger().Println(s.String())
			return
		}

		s := exitHandler(w, http.StatusOK, "Successfully restarted module [ %s ]", nil, M.Name())
		Agent.Logger().Println(s.String())
	})
}

// ReloadHandler will accept a request and attempt to
func reloadHandler(Agent *Agent) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		Name := mux.Vars(r)["Name"]

		M, err := Agent.GetByName(Name)
		if err != nil {
			s := exitHandler(w, http.StatusNotFound, "No module matching name [ %s ] could be found", err, Name)
			Agent.Logger().Println(s.String())
			return
		}

		if err := M.Reload(); err != nil {
			s := exitHandler(w, http.StatusInternalServerError, "Failed to reload module [ %s ]", err, M.Name())
			Agent.Logger().Println(s.String())
			return
		}

		s := exitHandler(w, http.StatusOK, "Successfully reloaded module [ %s ]", nil, M.Name())
		Agent.Logger().Println(s.String())
	})
}

// VersionHandler will accept a request and attempt to
func versionHandler(Agent *Agent) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		Name := mux.Vars(r)["Name"]

		M, err := Agent.GetByName(Name)
		if err != nil {
			s := exitHandler(w, http.StatusNotFound, "No module matching name [ %s ] could be found", err, Name)
			Agent.Logger().Println(s.String())
			return
		}

		V, err := M.GetVersion()
		if err != nil {
			s := exitHandler(w, http.StatusInternalServerError, "Failed to retrieve version for module [ %s ]", err, M.Name())
			Agent.Logger().Println(s.String())
			return
		}

		s := exitHandler(w, http.StatusOK, "Successfully retrieved version for module [ %s ] - %s", nil, M.Name(), V.String())
		Agent.Logger().Println(s.String())
	})
}

// StopHandler will accept a request and attempt to
func stopHandler(Agent *Agent) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Name := mux.Vars(r)["Name"]

		M, err := Agent.GetByName(Name)
		if err != nil {
			s := exitHandler(w, http.StatusNotFound, "No module matching name [ %s ] could be found", err, Name)
			Agent.Logger().Println(s.String())
			return
		}

		if err := M.Stop(); err != nil {
			s := exitHandler(w, http.StatusInternalServerError, "Failed to stop module [ %s ]", err, M.Name())
			Agent.Logger().Println(s.String())
			return
		}

		s := exitHandler(w, http.StatusOK, "Successfully stopped module [ %s ]", nil, M.Name())
		Agent.Logger().Println(s.String())
	})
}

// ActiveHandler will accept a request and attempt to
func activeHandler(Agent *Agent) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		b := bytes.NewBuffer(nil)

		for _, M := range Agent.Modules() {
			if M.State() == stateActive {
				b.WriteString(fmt.Sprintf("Module [ %s ] is [ %s ] (up for %s)\n", M.Name(), M.State(), M.Uptime().String()))
			}
		}

		if b.Len() == 0 {
			b.WriteString(fmt.Sprintf("No active modules found"))
		}

		exitHandler(w, http.StatusOK, b.String(), nil)

		for _, S := range strings.Split(b.String(), "\n") {
			if S != "" {
				Agent.Logger().Println(S)
			}
		}
	})
}

// StateHandler will accept a request and attempt to
func stateHandler(Agent *Agent) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Name := mux.Vars(r)["Name"]

		M, err := Agent.GetByName(Name)
		if err != nil {
			s := exitHandler(w, http.StatusNotFound, "No module matching name [ %s ] could be found", err, Name)
			Agent.Logger().Println(s.String())
			return
		}

		State := M.State()
		s := exitHandler(w, http.StatusOK, "Module [ %s ] is [ %s ] (up for %s)", nil, M.Name(), State.String(), M.Uptime())
		Agent.Logger().Println(s.String())
	})
}
