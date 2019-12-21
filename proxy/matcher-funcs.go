package proxy

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sort"
)

// ReverseProxyRouterFunc represents the Type which must be satisfied by any function which defines the per-request routing behaviours.  This must map a given request to a specific IP:Port host and leave the Path unchanged.
type ReverseProxyRouterFunc func(*http.Request) (Host string, Path string, err error)

// Define the set of errors provided by this package
var (
	ErrForbiddenRoute error = errors.New("easytls proxy error - Forbidden route")
	ErrRouteNotFound  error = errors.New("easytls routing rule error - No forwarding rule defined for route")
)

// LiveFileRouter implements a Reverse Proxy Routing function which will follow rules defined in a JSON file on disk. This rule-set is consulted on each incoming request, allowing any proxy using this to have the routing rules modified without an application restart.
func LiveFileRouter(RulesFilename string) ReverseProxyRouterFunc {
	return func(r *http.Request) (Host string, Path string, err error) {

		// Open the Rules file for reading.
		f, err := os.Open(RulesFilename)
		if os.IsNotExist(err) {
			return "", "", ErrRouteNotFound
		}
		if err != nil {
			return "", "", err
		}
		defer f.Close()

		// Read out the Rules
		RuleSet := ReverseProxyRuleSet{}
		if err := json.NewDecoder(f).Decode(&RuleSet); err != nil {
			return "", "", err
		}

		sort.Slice(RuleSet, RuleSet.Less)

		// Search for a match, and if one is found, define the new Host:Port based on what the rule determines.
		return RuleSet.Find(r.URL.EscapedPath())
	}
}

// DefinedRulesRouter will take in a pre-defined set of rules, and will route based on them.
//
// This may buy some efficiencies over the LiveFileRouter, as it doesn't need to perform Disk I/O on each request to search for the rules, but this comes with the tradeoff of not being able to edit the rules without restarting the application using this as the router.
func DefinedRulesRouter(RuleSet ReverseProxyRuleSet) ReverseProxyRouterFunc {

	sort.Slice(RuleSet, RuleSet.Less)

	return func(r *http.Request) (Host string, Path string, err error) {

		// Search for a match, and if one is found, define the new Host:Port based on what the rule determines.
		return RuleSet.Find(r.URL.EscapedPath())
	}
}
