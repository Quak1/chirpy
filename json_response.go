package main

import (
	"encoding/json"
	"log"
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

func respondJSONError(w http.ResponseWriter, statusCode int, msg string, err error) {
	type response struct {
		Error string `json:"error"`
	}

	if err != nil {
		log.Println(err)
	}

	respondJSON(w, statusCode, response{
		Error: msg,
	})
}
