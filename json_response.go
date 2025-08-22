package main

import (
	"encoding/json"
	"net/http"
)

func respondJSON(w http.ResponseWriter, statusCode int, resBody any) {
	data, err := json.Marshal(resBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)
}
