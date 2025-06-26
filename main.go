package main

import (
	"backend/config"
	"backend/handlers"
	"log"
	"net/http"
	"os"
)

func main() {
	// 确保音乐目录存在
	if _, err := os.Stat(config.MusicDir); os.IsNotExist(err) {
		log.Printf("Creat the music folder: %s", config.MusicDir)
		os.Mkdir(config.MusicDir, 0755)
	}

	// 创建静态文件服务
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	// API路由
	http.HandleFunc("/api/music", handlers.ListMusicHandler)
	http.HandleFunc("/api/stream/", handlers.StreamHandler)
	http.HandleFunc("/api/metadata/", handlers.MetadataHandler)
	http.HandleFunc("/api/lyric", handlers.GetLyricHandler)           // 获取歌词API
	http.HandleFunc("/api/upload-lyric", handlers.UploadLyricHandler) // 歌词上传API

	log.Printf("Server running on http://localhost%s", config.Port)
	log.Printf("At path: %s", config.MusicDir)
	log.Fatal(http.ListenAndServe(config.Port, nil))
}
