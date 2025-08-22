package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
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

	profanities := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := replaceProfaneWords(params.Body, profanities, "*")

	respondJSON(w, http.StatusOK, returnVals{
		CleanedBody: cleaned,
	})
}

func replaceProfaneWords(original string, profanities map[string]struct{}, c string) string {
	words := strings.Split(original, " ")

	for i, word := range words {
		_, isProfanity := profanities[strings.ToLower(word)]
		if isProfanity {
			words[i] = strings.Repeat(c, 4)
		}
	}

	return strings.Join(words, " ")
}
