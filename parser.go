package main

import (
	"encoding/json"
	"io"
	"net/http"
)

func FromJSON(r io.Reader, dst interface{}) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	return dec.Decode(dst)
}

func ToJSON(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
