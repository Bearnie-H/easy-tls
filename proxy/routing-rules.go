package proxy

import (
	"fmt"
	"net/url"
	"strings"
)

// ReverseProxyRoutingRule implements a single routing rule to be followed
// by the Reverse Proxy when re-routing traffic. This will take in a URL path,
// and return the Host:Port to forward the corresponding request to.
// This implementation is very basic, effectively effectively just re-routing
// to a new Host:Port based on the Path Prefix.
type ReverseProxyRoutingRule struct {
	PathPrefix      string
	DestinationHost string
	DestinationPort int
	NewPrefix       string
	ForbidRoute     bool
}

// ReverseProxyRuleSet implements a sortable interface for a set of
// Reverse Proxy Rules
type ReverseProxyRuleSet []ReverseProxyRoutingRule

func (a ReverseProxyRuleSet) Len() int      { return len(a) }
func (a ReverseProxyRuleSet) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less is defined in "reverse" order,
// as longer path-prefixes should be matched against first.
// This allows nested trees of URLs to potentially be proxied further
// to additional services or servers.
//
// For Example:
//
//	/foo/bar	-> Forward to service 1 and add a set of URL Query values
//	/foo		-> Forward to service 1 and do not add URL Query values
func (a ReverseProxyRuleSet) Less(i, j int) bool { return a[i].PathPrefix > a[j].PathPrefix }

// Find will return either the new Host:Port/Path to forward to
// or ErrRouteNotFound and nil
func (a ReverseProxyRuleSet) Find(in *url.URL) (out *url.URL, err error) {

	for _, Rule := range a {
		if Rule.matches(in) {
			return Rule.ToURL(in)
		}
	}

	return nil, ErrRouteNotFound
}

// Simple matching function, abstracted away to allow the "Rules"
// to become more complex as this library develops.
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

	switch {
	case (R.ForbidRoute):
		return fmt.Sprintf("Prefix: [ %s ] will forbid forwarding to [ %s:%d ].", R.PathPrefix, R.DestinationHost, R.DestinationPort)
	default:
		return fmt.Sprintf("Prefix: [ %s ] will forward to [ %s:%d ] and replace the prefix with [ %s ].", R.PathPrefix, R.DestinationHost, R.DestinationPort, R.NewPrefix)
	}
}

// ToURL will take in the incoming URL, and the rule it matches, and return
// a newly formatted URL with the modifications
func (R *ReverseProxyRoutingRule) ToURL(in *url.URL) (*url.URL, error) {

	if R.ForbidRoute {
		return nil, ErrForbiddenRoute
	}

	// Create a deep copy of the incoming URL
	out := &url.URL{}
	*out = *in

	// Set the host as per the rule.
	out.Host = fmt.Sprintf("%s:%d", R.DestinationHost, R.DestinationPort)

	// Replace the prefix with the specified value
	out.Path = strings.Replace(out.Path, R.PathPrefix, R.NewPrefix, 1)

	// If the scheme is empty, set to http
	if out.Scheme == "" {
		out.Scheme = "http"
	}

	// Other manipulations of the URI
	// ...

	return out, nil
}
