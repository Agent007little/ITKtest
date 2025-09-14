package responder

import (
	"encoding/json"
	"net/http"
)

type Responder interface {
	OutputJSON(w http.ResponseWriter, statusCode int, data interface{})
	Error(w http.ResponseWriter, statusCode int, message string)
}

type JSONResponder struct{}

func NewJSONResponder() *JSONResponder {
	return &JSONResponder{}
}

func (j *JSONResponder) OutputJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (j *JSONResponder) Error(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorResponse := map[string]string{
		"error": message,
	}
	
	json.NewEncoder(w).Encode(errorResponse)
}