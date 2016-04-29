package httpapi

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func makeErrorCheckedHandler(fn HttpHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s", err)
		}
	}
}

func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(makeErrorCheckedHandler(route.HandlerFunc))
	}

	return router
}
