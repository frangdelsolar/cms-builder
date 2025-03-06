package types

import "net/http"

// ApiHandlers holds the handlers for various API operations.
type ApiHandlers struct {
	List   ApiFunction
	Detail ApiFunction
	Create ApiFunction
	Update ApiFunction
	Delete ApiFunction
	Schema func(resource *Resource) http.HandlerFunc
}
