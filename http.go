package main

import (
	"github.com/capone212/supervisor2.prototype/httpapi"
	"log"
	"net/http"
)

func RunHttp() {

	router := httpapi.NewRouter()

	log.Fatal(http.ListenAndServe(":8085", router))
}
