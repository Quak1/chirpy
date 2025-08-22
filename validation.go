package main

import (
	"fmt"
	"strings"
)

func validateChirp(body string) (string, error) {
	if len(body) > 140 {
		return "", fmt.Errorf("Chirp is too long")
	}

	profanities := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := replaceProfaneWords(body, profanities, "*")
	return cleaned, nil
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
