package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func respondWithJSON(w http.ResponseWriter, code int, payload any) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) error {
	return respondWithJSON(w, code, map[string]string{"error": msg})
}

func replaceProfaneWords(replacableWords []string, replaceTo string, input string) string {
	separatedInput := strings.Split(input, " ")
	var output []string
	for _, inputWord := range separatedInput {
		written := false
		lowerInputWord := strings.ToLower(inputWord)
		for _, replacableWord := range replacableWords {
			if lowerInputWord == replacableWord {
				output = append(output, replaceTo)
				written = true
				continue
			}
		}
		if written {
			continue
		}
		output = append(output, inputWord)
	}
	return strings.Join(output, " ")
}
