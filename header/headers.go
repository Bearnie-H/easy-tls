package header

import (
	"net/http"
)

// Merge will add the contents of Insert into Base.
func Merge(Base *http.Header, Insert *http.Header) {
	for InsertKey, InsertValues := range *Insert {
		for _, InsertValue := range InsertValues {
			Base.Add(InsertKey, InsertValue)
		}
	}
}
