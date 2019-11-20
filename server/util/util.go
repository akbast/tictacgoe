package util

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteResponse(response interface{}, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	fmt.Println("Formatter called")
	json.NewEncoder(w).Encode(response)
}

func FormatMessageAsJson(message string) map[string]interface{} {
	return map[string]interface{}{"message": message}
}
