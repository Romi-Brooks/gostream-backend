package handlers

import (
	"backend/config"
	"backend/models"
	"backend/utils"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func MetadataHandler(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)

	filePath := strings.TrimPrefix(r.URL.Path, "/api/metadata/")

	// 防止路径遍历攻击
	if strings.Contains(filePath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(config.MusicDir, filePath)
	log.Printf("Get file's metadata: %s", fullPath)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		log.Printf("File does not exist: %s", fullPath)
		http.NotFound(w, r)
		return
	}

	// 获取元数据
	metadata, err := utils.GetMusicMetadata(fullPath)
	if err != nil {
		log.Printf("Error when get file's metadata: %v", err)
		// 返回默认元数据
		info, _ := os.Stat(fullPath)
		response := models.MusicMetadata{
			Title:  strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())),
			Artist: "未知艺术家",
			Album:  "未知专辑",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}
