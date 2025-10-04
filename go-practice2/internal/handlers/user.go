package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func UserHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		handleGetUser(w, r)
		return
	}

	if r.Method == http.MethodPost {
		handlePostUser(w, r)
		return
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	response := map[string]int{"user_id": id}
	json.NewEncoder(w).Encode(response)
}

func handlePostUser(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Name string `json:"name"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil || data.Name == "" {
		http.Error(w, `{"error":"invalid name"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	response := map[string]string{"created": data.Name}
	json.NewEncoder(w).Encode(response)
}
