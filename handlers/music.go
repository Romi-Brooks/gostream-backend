package handlers

import (
	"backend/config"
	"backend/utils"
	"encoding/json"
	"log"
	"net/http"
)

func EnableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func ListMusicHandler(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	files, err := utils.GetMusicFiles(config.MusicDir)
	if err != nil {
		log.Printf("Error when get the file list: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}
