package customerrors

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteError(w http.ResponseWriter, err error) {
	customErr := Resolve(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(customErr.StatusCode)

	if encErr := json.NewEncoder(w).Encode(customErr); encErr != nil {
		// headers already flushed; status can't change but log so the operator sees it
		log.Printf("customerrors: failed to encode error response: %v", encErr)
	}
}
