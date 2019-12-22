package common

import "net/http"

// AddHeaders will add a Header map into the given HTTP Header.
func AddHeaders(Header *http.Header, ToAdd map[string][]string) {
	for key, values := range ToAdd {
		for _, value := range values {
			Header.Add(key, value)
		}
	}
}
