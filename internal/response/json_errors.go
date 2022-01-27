package response

import (
	"encoding/json"
	"log"
	"net/http"
)

type jsonErrors struct {
	Message string  `json:"message"`
	Errors  []Error `json:"errors"`
}

type Error struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func WriteJSONErrors(w http.ResponseWriter, message string, errors []Error, statusCode int) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(jsonErrors{
		Message: message,
		Errors:  errors,
	}); err != nil {
		log.Println("handler/middleware failed to write response:", err)
	}
}

func WriteJSONError(w http.ResponseWriter, name string, description string, statusCode int) {
	WriteJSONErrors(w, name, []Error{{Name: name, Description: description}}, statusCode)
}
