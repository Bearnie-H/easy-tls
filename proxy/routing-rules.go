package proxy

import (
	"fmt"
	"net/url"
	"strings"
)

// ReverseProxyRoutingRule implements a single routing rule to be followed by the Reverse Proxy when re-routing traffic.  This will take in a URL path, and return the Host:Port to forward the corresponding request to.  This implementation is very basic, effectively effectively just re-routing to a new Host:Port based on the Path Prefix.
type ReverseProxyRoutingRule struct {
	PathPrefix      string
	DestinationHost string
	DestinationPort int
	StripPrefix     bool
	ForbidRoute     bool
}

// ReverseProxyRuleSet implements a sortable interface for a set of Reverse Proxy Rules
type ReverseProxyRuleSet []ReverseProxyRoutingRule

func (a ReverseProxyRuleSet) Len() int      { return len(a) }
func (a ReverseProxyRuleSet) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less is defined in "reverse" order, as longer path-prefixes should be matched against first.
func (a ReverseProxyRuleSet) Less(i, j int) bool { return a[i].PathPrefix > a[j].PathPrefix }

// Find will return either the new Host:Port/Path to forward to, or ErrRouteNotFound
func (a ReverseProxyRuleSet) Find(in *url.URL) (out *url.URL, err error) {
	for _, Rule := range a {
		if Rule.matches(in) {
			if Rule.ForbidRoute {
				return nil, ErrForbiddenRoute
			}
			return Rule.ToURL(in)
		}
	}

	return nil, ErrRouteNotFound
}

// Simple matching function, abstracted away to allow the "Rules" to become more complex as this library develops.
func (R *ReverseProxyRoutingRule) matches(in *url.URL) bool {

	// Check if the paths match...

	if !strings.HasPrefix(in.Path, "/") {
		in.Path = "/" + in.Path
	}

	if !strings.HasPrefix(R.PathPrefix, "/") {
		R.PathPrefix = "/" + R.PathPrefix
	}

	// Define other "match" criteria
	// ...

	return strings.HasPrefix(in.Path, R.PathPrefix)
}

func (R *ReverseProxyRoutingRule) String() string {

	var key uint64 = 0
	// Set Bit 0 if the rule should strip the prefix.
	if R.StripPrefix {
		key += 1 << 0
	}
	// Set Bit 1 if the rule should forbid traffic on the route.
	if R.ForbidRoute {
		key += 1 << 1
	}

	switch key {
	case 0:
		return fmt.Sprintf("Prefix: [ %s ] will forward to [ %s:%d ] without stripping the prefix.", R.PathPrefix, R.DestinationHost, R.DestinationPort)
	case 1:
		return fmt.Sprintf("Prefix: [ %s ] will forward to [ %s:%d ] while stripping the prefix.", R.PathPrefix, R.DestinationHost, R.DestinationPort)
	case 2:
		return fmt.Sprintf("Prefix: [ %s ] will forbid forwarding to [ %s:%d ] without stripping the prefix.", R.PathPrefix, R.DestinationHost, R.DestinationPort)
	case 3:
		return fmt.Sprintf("Prefix: [ %s ] will forbid forwarding to [ %s:%d ] while stripping the prefix.", R.PathPrefix, R.DestinationHost, R.DestinationPort)
	default:
		return fmt.Sprintf("Prefix: [ %s ] contains an unknown combination of flags.", R.PathPrefix)
	}
}

// ToURL will take in the incoming URL, and the rule it matches, and return a newly formatted URL with the modifications
func (R *ReverseProxyRoutingRule) ToURL(in *url.URL) (*url.URL, error) {

	// Create a deep copy of the incoming URL
	out := &url.URL{}
	*out = *in

	// Set the host as per the rule.
	out.Host = fmt.Sprintf("%s:%d", R.DestinationHost, R.DestinationPort)

	// Strip the prefix if specified by the rule.
	if R.StripPrefix {
		out.Path = strings.TrimPrefix(out.Path, R.PathPrefix)
	}

	// More manipulations of the incoming URL before returning it
	// ...

	return out, nil
}
