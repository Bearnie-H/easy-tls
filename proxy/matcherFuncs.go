package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// ReverseProxyRouterFunc represents the Type which must be satisfied by any function which defines the per-request routing behaviours.  This must map a given request to a specific IP:Port host.
type ReverseProxyRouterFunc func(*http.Request) (string, error)

// ReverseProxyRoutingRule implements a single routing rule to be followed by the Reverse Proxy when re-routing traffic.  This will take in a URL path, and return the Host:Port to forward the corresponding request to.  This implementation is very basic, effectively effectively just re-routing to a new Host:Port based on the Path Prefix.
type ReverseProxyRoutingRule struct {
	PathPrefix      string
	DestinationHost string
	DestinationPort int
}

// reverseProxyRoutingRuleSet aliases an iterable set of routing rules.
type reverseProxyRoutingRuleSet []ReverseProxyRoutingRule

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

// LiveFileRouter implements a Reverse Proxy Routing function which will follow rules defined in a JSON file on disk. This rule-set is consulted on each incoming request,
func LiveFileRouter(RulesFilename string) ReverseProxyRouterFunc {
	return func(r *http.Request) (string, error) {

		// Open the Rules file for reading.
		f, err := os.Open(RulesFilename)
		if err != nil {
			return "", err
		}
		defer f.Close()

		// Read out the Rules
		RuleSet := reverseProxyRoutingRuleSet{}
		if err := json.NewDecoder(f).Decode(&RuleSet); err != nil {
			return "", err
		}

		// Search for a match, and if one is found, define the new Host:Port based on what the rule determines.
		for _, Rule := range RuleSet {
			if Rule.matches(r.URL.EscapedPath()) {
				return fmt.Sprintf("%s:%d", Rule.DestinationHost, Rule.DestinationPort), nil
			}
		}

		return "", fmt.Errorf("reverse proxy error - no defined rule for path %s", r.URL.EscapedPath())
	}
}
