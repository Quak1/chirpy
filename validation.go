package main

import (
	"encoding/json"
	"net/http"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		Valid bool `json:"valid"`
	}
	type errorRes struct {
		Error string `json:"error"`
	}

	params := parameters{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(params.Body) > 140 {
		respondJSON(w, http.StatusBadRequest, errorRes{
			Error: "Chirp is too lonng",
		})
		return
	}

	respondJSON(w, http.StatusOK, returnVals{
		Valid: true,
	})
}
