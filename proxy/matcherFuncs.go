package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// ReverseProxyRouterFunc represents the Type which must be satisfied by any function which defines the per-request routing behaviours.  This must map a given request to a specific IP:Port host and leave the Path unchanged.
type ReverseProxyRouterFunc func(*http.Request) (string, error)

// ErrForbiddenRoute represents the error to return when a proxy finds a route which is forbidden.
var ErrForbiddenRoute error = errors.New("easytls proxy error - Forbidden route")

// ErrRouteNotFound is the error code returned when a route cannot be found in the set of given Rules
var ErrRouteNotFound error = errors.New("easytls routing rule error - No rule defined for route")

// ReverseProxyRoutingRule implements a single routing rule to be followed by the Reverse Proxy when re-routing traffic.  This will take in a URL path, and return the Host:Port to forward the corresponding request to.  This implementation is very basic, effectively effectively just re-routing to a new Host:Port based on the Path Prefix.
type ReverseProxyRoutingRule struct {
	PathPrefix      string
	DestinationHost string
	DestinationPort int
	StripPrefix     bool
}

// Simple matching function, abstracted away to allow the "Rules" to become more complex as this library develops.
func (R *ReverseProxyRoutingRule) matches(Path string) bool {

	if !strings.HasPrefix(Path, "/") {
		Path = "/" + Path
	}

	if !strings.HasPrefix(R.PathPrefix, "/") {
		R.PathPrefix = "/" + R.PathPrefix
	}

	return strings.HasPrefix(Path, R.PathPrefix)
}

func (R *ReverseProxyRoutingRule) String() string {
	if R.StripPrefix {
		return fmt.Sprintf("Prefix: \"%s\" will forward to \"%s:%d\" while stripping the prefix.", R.PathPrefix, R.DestinationHost, R.DestinationPort)
	}
	return fmt.Sprintf("Prefix: \"%s\" will forward to \"%s:%d\" without stripping the prefix.", R.PathPrefix, R.DestinationHost, R.DestinationPort)
}

// ToURL will convert a URL path, using the Rule to build a Host:Port/Path URL string.
func (R *ReverseProxyRoutingRule) ToURL(Path string) string {
	if strings.HasSuffix(R.PathPrefix, "/") {
		R.PathPrefix = strings.TrimSuffix(R.PathPrefix, "/")
	}
	if R.StripPrefix {
		return fmt.Sprintf("%s:%d/%s", R.DestinationHost, R.DestinationPort, strings.TrimPrefix(Path, R.PathPrefix))
	}
	return fmt.Sprintf("%s:%d%s", R.DestinationHost, R.DestinationPort, Path)
}

// LiveFileRouter implements a Reverse Proxy Routing function which will follow rules defined in a JSON file on disk. This rule-set is consulted on each incoming request, allowing any proxy using this to have the routing rules modified without an application restart.
func LiveFileRouter(RulesFilename string) ReverseProxyRouterFunc {
	return func(r *http.Request) (string, error) {

		// Open the Rules file for reading.
		f, err := os.Open(RulesFilename)
		if os.IsNotExist(err) {
			return "", ErrRouteNotFound
		}
		if err != nil {
			return "", err
		}
		defer f.Close()

		// Read out the Rules
		RuleSet := []ReverseProxyRoutingRule{}
		if err := json.NewDecoder(f).Decode(&RuleSet); err != nil {
			return "", err
		}

		// Search for a match, and if one is found, define the new Host:Port based on what the rule determines.
		for _, Rule := range RuleSet {
			if Rule.matches(r.URL.EscapedPath()) {
				return Rule.ToURL(r.URL.EscapedPath()), nil
			}
		}

		return "", ErrRouteNotFound
	}
}

// DefinedRulesRouter will take in a pre-defined set of rules, and will route based on them.
//
// This may buy some efficiencies over the LiveFileRouter, as it doesn't need to perform Disk I/O on each request to search for the rules, but this comes with the tradeoff of not being able to edit the rules without restarting the application using this as the router.
func DefinedRulesRouter(RuleSet []ReverseProxyRoutingRule) ReverseProxyRouterFunc {
	return func(r *http.Request) (string, error) {
		// Search for a match, and if one is found, define the new Host:Port based on what the rule determines.
		for _, Rule := range RuleSet {
			if Rule.matches(r.URL.EscapedPath()) {
				return fmt.Sprintf("%s:%d", Rule.DestinationHost, Rule.DestinationPort), nil
			}
		}

		return "", ErrRouteNotFound
	}
}
