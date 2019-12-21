package proxy

import (
	"fmt"
	"strings"
)

// ReverseProxyRoutingRule implements a single routing rule to be followed by the Reverse Proxy when re-routing traffic.  This will take in a URL path, and return the Host:Port to forward the corresponding request to.  This implementation is very basic, effectively effectively just re-routing to a new Host:Port based on the Path Prefix.
type ReverseProxyRoutingRule struct {
	PathPrefix      string
	DestinationHost string
	DestinationPort int
	StripPrefix     bool
}

// ReverseProxyRuleSet implements a sortable interface for a set of Reverse Proxy Rules
type ReverseProxyRuleSet []ReverseProxyRoutingRule

func (a ReverseProxyRuleSet) Len() int           { return len(a) }
func (a ReverseProxyRuleSet) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ReverseProxyRuleSet) Less(i, j int) bool { return a[i].PathPrefix < a[j].PathPrefix }

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
func (R *ReverseProxyRoutingRule) ToURL(PathIn string) (string, string) {
	if strings.HasSuffix(R.PathPrefix, "/") {
		R.PathPrefix = strings.TrimSuffix(R.PathPrefix, "/")
	}
	if R.StripPrefix {
		return fmt.Sprintf("%s:%d", R.DestinationHost, R.DestinationPort), strings.TrimPrefix(PathIn, R.PathPrefix)
	}
	return fmt.Sprintf("%s:%d", R.DestinationHost, R.DestinationPort), PathIn
}
