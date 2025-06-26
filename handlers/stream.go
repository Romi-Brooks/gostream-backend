package handlers

import (
	"backend/config"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func StreamHandler(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)

	filePath := strings.TrimPrefix(r.URL.Path, "/api/stream/")
	log.Printf("Get the file: %s", filePath)

	// 防止路径遍历攻击
	if strings.Contains(filePath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(config.MusicDir, filePath)
	log.Printf("Full path: %s", fullPath)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		log.Printf("File does not exist: %s", fullPath)
		http.NotFound(w, r)
		return
	}

	// 设置正确的Content-Type
	ext := strings.ToLower(filepath.Ext(fullPath))
	switch ext {
	case ".mp3":
		w.Header().Set("Content-Type", "audio/mpeg")
	case ".flac":
		w.Header().Set("Content-Type", "audio/flac")
	case ".wav":
		w.Header().Set("Content-Type", "audio/wav")
	case ".ogg":
		w.Header().Set("Content-Type", "audio/ogg")
	case ".m4a":
		w.Header().Set("Content-Type", "audio/mp4")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// 支持断点续传
	http.ServeFile(w, r, fullPath)
}
