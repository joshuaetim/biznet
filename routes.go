package main

import "net/http"

func RegisterRoutes(base *http.ServeMux, handler *RecordHandler) {
	base.HandleFunc("POST /records", handler.Create)
	base.HandleFunc("PUT /records/{id}", handler.Update)
	base.HandleFunc("DELETE /records/{id}", handler.Delete)
}
